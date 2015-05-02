package server

import (
	"github.com/koding/kite"
	"log"
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
}

func NewBroker(dbpath string, k *kite.Kite) *Broker {

	b := &Broker{
		ID:      1,
		DBPath:  dbpath,
		Topics:  make(map[string]*Topic),
		Workers: make(map[string]*Worker),
		idChan:  make(chan MessageID, 4096), // Buffer
		kite:    k,
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
		logger.Error("db path", "err", err)
		return err
	}

	for _, p := range dbfilepaths {
		topicName := strings.Replace(filepath.Base(p), filepath.Ext(p), "", 1)
		b.Topic(topicName)
		logger.Info("Load topic db", "topic", topicName, "db", p)
	}

	//Register Worker RPC
	b.kite.HandleFunc("loom.server:worker.connect", b.HandleWorkerConnect)
	b.kite.HandleFunc("loom.server:worker.job", b.HandleWorkerJobFeedback)
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

	logger.Info("worker.init", "worker", workerId, "topic", topicName)
	return true, nil
}

func (b *Broker) HandleWorkerJobFeedback(r *kite.Request) (interface{}, error) {
	args := r.Args.MustSlice()
	msgIdStr := args[0].MustString()
	topic := args[1].MustString()
	state := args[2].MustString()

	if state == "done" {
		topic := b.Topic(topic)
		var msgId MessageID
		copy(msgId[:], []byte(msgIdStr))
		err := topic.FinishMessage(msgId)
		if err != nil {
			logger.Error("HandleWorkerJobFeedback", "err", err, "id", string(msgId[:]))
		}
	}

	return nil, nil
}

func (b *Broker) WorkerDisconnect(c *kite.Client) {
	b.workersMutex.Lock()
	defer b.workersMutex.Unlock()

	if w, ok := b.Workers[c.ID]; ok {
		topic := b.Topic(w.TopicName)
		topic.Dispatcher.RemoveWorker(w)
		delete(b.Workers, c.ID)
		logger.Info("worker.disconnect", "worker", c.ID)
	}
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
		logger.Error("Load topic db", "err", err)
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
				log.Printf("ERROR: %s", err)

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
