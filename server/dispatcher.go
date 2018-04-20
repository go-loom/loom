package server

import (
	"github.com/go-loom/loom/log"
	"sync"
)

type Dispatcher struct {
	topicName    string
	workers      map[string]*Worker
	workersMutex sync.RWMutex
	msgPopChan   chan *Message
	msgPushChan  chan *Message
	quitChan     chan struct{}
	store        Store
	running      bool
	logger       log.Logger
}

func NewDispatcher(topicName string, store Store) *Dispatcher {
	d := &Dispatcher{
		topicName:   topicName,
		workers:     make(map[string]*Worker),
		msgPopChan:  make(chan *Message),
		msgPushChan: make(chan *Message),
		quitChan:    make(chan struct{}),
		store:       store,
		logger:      log.New("Dispatcher:" + topicName),
	}

	return d
}

func (d *Dispatcher) AddWorker(w *Worker) {
	d.workersMutex.Lock()
	defer d.workersMutex.Unlock()
	d.workers[w.ID] = w
	w.SetDispatcher(d)
	if d.running == false {
		go d.dispatching()
		d.running = true
	}
	d.logger.Info("AddWorker workers:%v running:%v", len(d.workers), d.running)
}

func (d *Dispatcher) RemoveWorker(w *Worker) {
	d.workersMutex.Lock()
	defer d.workersMutex.Unlock()

	delete(d.workers, w.ID)

	if len(d.workers) <= 0 && d.running == true {
		d.quitChan <- struct{}{}
		d.running = false
	}

	d.logger.Info("RemoveWorker workers:%v running:%v", len(d.workers), d.running)
}

func (d *Dispatcher) NotifyWorkerState() {
	d.workersMutex.Lock()
	defer d.workersMutex.Unlock()

	num := 0
	for _, w := range d.workers {
		if w.Working() {
			num++
		}
	}
	d.logger.Info("Notify num:%v running:%v", num, d.running)

	if num == 0 && d.running == true {
		d.quitChan <- struct{}{}
		d.running = false
	} else if num > 0 && d.running == false {
		go d.dispatching()
		d.running = true
	}
}

func (d *Dispatcher) dispatching() {
	d.logger.Info("Start dispatching")

L:
	for {
		select {
		case <-d.quitChan:
			d.logger.Info("Recv quit dispatching")
			break L
		case msg := <-d.msgPopChan:
			d.logger.Info("pop msg")

			workers := d.selectWorkers()
			if len(workers) <= 0 {
				d.msgPushChan <- msg
				d.logger.Info("No running workers. repush msg id:%v", string(msg.ID[:]))
			}

			d.logger.Info("Running workers:%v", len(workers))

			for _, w := range workers {
				ok := w.SendMessage(msg)

				if !ok {
					d.logger.Error("Sending message is failure id:%v", string(msg.ID[:]))
					d.msgPushChan <- msg
				} else {
					msg.State = MSG_RECEIVED

					b := d.store.MessageBucket(WorkerPerMessageBucketNamePrefix + w.ID)
					err := b.Put(msg)
					if err != nil {
						d.logger.Error("Save pending msg per worker id:%v err:%v", string(msg.ID[:]), err)
					}

					err = d.store.MessageBucket(MessageBucketName).Put(msg)
					if err != nil {
						d.logger.Error("Save msg id:%v err:%v", string(msg.ID[:]), err)
					}

					d.logger.Info("Pop message id:%v", string(msg.ID[:]))
				}
			}
			d.logger.Info("E: pop msg id:%v", string(msg.ID[:]))
		}
	}

	d.logger.Info("End dispatching")
}

func (d *Dispatcher) selectWorkers() []*Worker {
	d.logger.Info("selectWorkers..")
	var ws []*Worker
	for _, w := range d.workers {
		if w.Working() {
			ws = append(ws, w)
			break
		}
	}

	d.logger.Info("E:selectWorkers..")
	return ws
}
