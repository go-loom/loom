package server

import (
	"github.com/koding/kite"
	"golang.org/x/net/context"
	"gopkg.in/loom.v1/config"
	"gopkg.in/loom.v1/log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Broker struct {
	ctx          context.Context
	ID           int64
	DBPath       string
	Topics       map[string]*Topic
	topicMutex   sync.Mutex
	Workers      map[string]*Worker
	workersMutex sync.RWMutex
	kite         *kite.Kite
	idChan       chan MessageID
	topicQuitC   chan struct{}
	quitC        chan struct{}
	wg           sync.WaitGroup
	logger       log.Logger
}

func NewBroker(ctx context.Context, dbpath string, k *kite.Kite) *Broker {

	b := &Broker{
		ctx:        ctx,
		ID:         1,
		DBPath:     dbpath,
		Topics:     make(map[string]*Topic),
		Workers:    make(map[string]*Worker),
		idChan:     make(chan MessageID, 4096), // Buffer
		topicQuitC: make(chan struct{}),
		quitC:      make(chan struct{}),
		kite:       k,
		logger:     log.New("Broker"),
	}

	go b.recvTopicQuit()
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
	b.kite.HandleFunc("loom.server:worker.info", b.HandleWorkerInfo)
	b.kite.OnDisconnect(b.WorkerDisconnect)

	return nil
}

func (b *Broker) Done() {
	b.wg.Wait()
	b.logger.Info("Closing broker..")
	return
}

func (b *Broker) HandleWorkerConnect(r *kite.Request) (interface{}, error) {
	args := r.Args.MustSlice()

	workerId := args[0].MustString()
	topicName := args[1].MustString()
	maxJobSize := int(args[2].MustFloat64())
	r.Client.ID = workerId

	b.workersMutex.Lock()
	defer b.workersMutex.Unlock()

	w := NewConnectedWorker(workerId, topicName, maxJobSize, r.Client)

	b.Workers[workerId] = w

	topic := b.Topic(topicName)

	topic.Dispatcher.AddWorker(w)

	b.logger.Info("Topic#%v/Worker#%v connected (maxJobSize:%v)", workerId, topicName, maxJobSize)
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

	pmb := topic.store.MessageBucket(WorkerPerMessageBucketNamePrefix + r.Client.ID)
	err = pmb.Del(msgId)
	if err != nil {
		b.logger.Error("Remove pending msg worker:%v messageId:%v err:%v", r.Client.ID, msgIdStr, err)
	}

	b.logger.Info("Finish message id:%v from worker:%v", msgIdStr, r.Client.ID)
	return nil, nil
}

func (b *Broker) HandleWorkerJobTasksState(r *kite.Request) (interface{}, error) {

	args := r.Args.MustSlice()
	var tasks *map[string]interface{}
	args[0].MustUnmarshal(&tasks)
	msgIdStr := args[1].MustString()
	topicName := args[2].MustString()

	topic := b.Topic(topicName)
	msgId := GetMessageID([]byte(msgIdStr))

	msg, err := topic.msgBucket.Get(msgId)
	msg.State = MSG_RECEIVED
	if err != nil {
		b.logger.Error("Save job task results id:%v err:%v", msgIdStr, err)
		return nil, nil
	}
	if msg == nil {
		b.logger.Error("Save job task results id:%v err:%v", msgIdStr, "msg is empty")
		return nil, nil
	}

	msg.SetResults(r.Client.ID, *tasks)
	err = topic.msgBucket.Put(msg)
	if err != nil {
		b.logger.Error("Save job task results id:%v err: %v", msgIdStr, err)
	}

	b.logger.Info("Received task results id:%v from worker:%v", msgIdStr, r.Client.ID)

	return nil, nil
}

func (b *Broker) HandleWorkerInfo(r *kite.Request) (interface{}, error) {
	args := r.Args.MustSlice()
	numJob := int(args[0].MustFloat64())
	workerId := r.Client.ID

	w := b.getWorker(workerId)
	w.SetNumJob(numJob)

	b.logger.Info("Received worker info worker:%v jobs:%v", workerId, numJob)
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

//TODO: It's not used
func (b *Broker) HandleWorkerShutdown(r *kite.Request) (interface{}, error) {
	w := b.getWorker(r.Client.ID)

	if w != nil {
		w.SetWorking(false)

	}

	b.logger.Info("Worker:%v is shutdowning.", r.Client.ID)
	return nil, nil
}

func (b *Broker) getWorker(workerId string) *Worker {
	b.workersMutex.RLock()
	defer b.workersMutex.RUnlock()

	if w, ok := b.Workers[workerId]; ok {
		return w
	}
	return nil
}

func (b *Broker) Topic(name string) *Topic {
	b.topicMutex.Lock()
	defer b.topicMutex.Unlock()

	if t, ok := b.Topics[name]; ok {
		return t
	}

	topicCtx := context.WithValue(b.ctx, "quitC", b.topicQuitC)

	store, _ := NewTopicStore("bolt", b.DBPath, name)
	store.Open()

	pendingTimeout := 10 * time.Second
	t := NewTopic(topicCtx, name, pendingTimeout, store)
	err := t.Init()
	if err != nil {
		b.logger.Error("Load topic %v db err: %v", name, err)
	}

	b.Topics[name] = t

	b.wg.Add(1) //Add topic wait group count

	return t
}

func (b *Broker) PushMessage(name string, job *config.Job) (*Message, error) {
	t := b.Topic(name)
	msg := NewMessage(b.NewID(), job)
	t.PushMessage(msg)
	return msg, nil
}

func (b *Broker) GetMessage(name string, id MessageID) (*Message, error) {
	t := b.Topic(name)

	msg, err := t.msgBucket.Get(id)
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

func (b *Broker) recvTopicQuit() {
	for _ = range b.topicQuitC {
		b.wg.Done()
	}
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
		case <-b.ctx.Done():
			break
		}
	}

	b.logger.Info("End idPump")
}
