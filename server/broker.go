package server

import (
	"github.com/koding/kite"
	"gopkg.in/loom.v1/log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Broker struct {
	ID           int64
	DBPath       string
	Topics       map[string]*Topic
	topicMutex   sync.Mutex
	Workers      map[string]*Worker
	workersMutex sync.RWMutex
	kite         *kite.Kite
	idChan       chan MessageID
	logger       log.Logger
}

func NewBroker(dbpath string, k *kite.Kite) *Broker {

	b := &Broker{
		ID:      1,
		DBPath:  dbpath,
		Topics:  make(map[string]*Topic),
		Workers: make(map[string]*Worker),
		idChan:  make(chan MessageID, 4096), // Buffer
		kite:    k,
		logger:  log.New("Broker"),
	}

	return b
}

func (b *Broker) Init() error {
	go b.idPump()

	if _, err := os.Stat(b.DBPath); os.IsNotExist(err) {
		err := os.MkdirAll(b.DBPath, 0777)
		if err != nil {
			return err
		}
	}

	pattern := filepath.Join(b.DBPath, "*.boltdb")
	dbfilepaths, err := filepath.Glob(pattern)
	if err != nil {
		b.logger.Error("Wrong dbpath err: %v", err)
		return err
	}

	for _, p := range dbfilepaths {
		topicName := strings.Replace(filepath.Base(p), filepath.Ext(p), "", 1)
		b.Topic(topicName)
		b.logger.Info("load topic db topic: %v db: %v", topicName, p)
	}

	//Register Worker RPC
	b.kite.HandleFunc("loom.server:worker.connect", b.HandleWorkerConnect)
	b.kite.HandleFunc("loom.server:worker.job.tasks.state", b.HandleWorkerJobTasksState)
	b.kite.HandleFunc("loom.server:worker.job.done", b.HandleWorkerJobDone)
	b.kite.HandleFunc("loom.server:worker.shutdown", b.HandleWorkerShutdown)
	b.kite.OnDisconnect(b.WorkerDisconnect)

	return nil
}

func (b *Broker) HandleWorkerConnect(r *kite.Request) (interface{}, error) {
	args := r.Args.MustSlice()

	workerId := args[0].MustString()
	topicName := args[1].MustString()

	r.Client.ID = workerId

	b.workersMutex.Lock()
	defer b.workersMutex.Unlock()

	w := NewConnectedWorker(workerId, topicName, r.Client)
	b.Workers[workerId] = w

	topic := b.Topic(topicName)
	topic.Dispatcher.AddWorker(w)

	b.logger.Info("Worker %v of topic %v connected", workerId, topicName)
	return true, nil
}

func (b *Broker) HandleWorkerJobDone(r *kite.Request) (interface{}, error) {
	args := r.Args.MustSlice()
	msgIdStr := args[0].MustString()
	topicName := args[1].MustString()

	topic := b.Topic(topicName)
	var msgId MessageID
	copy(msgId[:], []byte(msgIdStr))

	err := topic.FinishMessage(msgId)
	if err != nil {
		b.logger.Error("Finish message messageId:%v err:%v", msgIdStr, err)
	}

	b.logger.Info("Finish message id:%v from worker:%v", msgIdStr, r.Client.ID)
	return nil, nil
}

func (b *Broker) HandleWorkerJobTasksState(r *kite.Request) (interface{}, error) {
	args := r.Args.MustSlice()
	var tasks *[]map[string]interface{}
	args[0].MustUnmarshal(&tasks)
	msgIdStr := args[1].MustString()
	topicName := args[2].MustString()

	topic := b.Topic(topicName)
	var msgId MessageID
	copy(msgId[:], []byte(msgIdStr))

	err := topic.store.SaveTasks(msgId, r.Client.ID, *tasks)
	if err != nil {
		b.logger.Error("Save job feedback messageId: %v err: %v", msgIdStr, err)
	}

	b.logger.Info("Received job:%v from worker:%v", msgIdStr, r.Client.ID)
	return nil, nil
}

func (b *Broker) WorkerDisconnect(c *kite.Client) {
	b.workersMutex.Lock()
	defer b.workersMutex.Unlock()

	//Msgs which is working on the worker (not finished jobs) should be enqueued again.
	if w, ok := b.Workers[c.ID]; ok {
		topic := b.Topic(w.TopicName)
		topic.Dispatcher.RemoveWorker(w)
		topic.PushPendingMsgsInWorker(w.ID)

		delete(b.Workers, c.ID)
		b.logger.Info("Worker %s disconneced", c.ID)
	}
}

func (b *Broker) HandleWorkerShutdown(r *kite.Request) (interface{}, error) {
	b.workersMutex.RLock()
	defer b.workersMutex.RUnlock()

	//The worker can't be working anymore.
	if w, ok := b.Workers[r.Client.ID]; ok {
		w.SetWorking(false)
	}
	b.logger.Info("Worker:%v is shutdowning.", r.Client.ID)
	return nil, nil
}

func (b *Broker) Topic(name string) *Topic {
	b.topicMutex.Lock()
	defer b.topicMutex.Unlock()

	if t, ok := b.Topics[name]; ok {
		return t
	}

	store, _ := NewTopicStore("bolt", b.DBPath, name)
	store.Open()

	pendingTimeout := 10 * time.Second
	t := NewTopic(name, pendingTimeout, store)
	err := t.Init()
	if err != nil {
		b.logger.Error("Load topic %v db err: %v", name, err)
	}

	b.Topics[name] = t

	return t
}

func (b *Broker) PushMessage(name string, value interface{}) (*Message, error) {
	t := b.Topic(name)
	msg := NewMessage(b.NewID(), value.([]byte))
	t.PushMessage(msg)
	return msg, nil
}

func (b *Broker) GetMessage(name string, id MessageID) (*Message, error) {
	t := b.Topic(name)

	msg, err := t.store.GetMessage(id)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (b *Broker) FinishMessage(name string, id MessageID) error {
	t := b.Topic(name)
	err := t.FinishMessage(id)
	return err
}

//This method is implemented to MessageIDGenerator
func (b *Broker) NewID() MessageID {
	return <-b.idChan
}

func (b *Broker) idPump() {
	factory := &guidFactory{}
	lastError := time.Now()

	for {

		id, err := factory.NewGUID(b.ID)

		if err != nil {
			now := time.Now()
			if now.Sub(lastError) > time.Second {
				//only print the error once/second
				//TODO:
				lastError = now
				b.logger.Error("id pump err: %s", err)

			}
			runtime.Gosched()
			continue
		}

		select {
		case b.idChan <- id.Hex():
			//TODO: exitChan?
		}
	}
}
