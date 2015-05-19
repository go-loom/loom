package server

import (
	"gopkg.in/loom.v1/log"
	"sync"
	"time"
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
	d.logger.Info("notify worker state")

	d.workersMutex.RLock()
	defer d.workersMutex.RUnlock()

	num := 0
	for _, w := range d.workers {
		if w.Working() {
			num++
		}
	}
	d.logger.Info("notify num:%v running:%v", num, d.running)

	if num == 0 && d.running == true {
		d.quitChan <- struct{}{}
		d.running = false
		d.logger.Info("Stop")
	} else if d.running == false {
		go d.dispatching()
		d.running = true
		d.logger.Info("Start")
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
				d.logger.Info("msgPushChan")
				d.msgPushChan <- msg

				d.logger.Info("msgPushChan")
			}
			d.logger.Info("Dispatched Workers:%v", len(workers))
			for _, w := range workers {
				d.logger.Info("S:Sent message %v", string(msg.ID[:]))
				ok := w.SendMessage(msg)
				d.logger.Info("E:Sent message %v", string(msg.ID[:]))

				if !ok {
					d.logger.Error("Sending message is failure id:%v", string(msg.ID[:]))
					d.msgPushChan <- msg
				} else {
					pendingMsg := &PendingMessage{
						MessageID: msg.ID,
						WorkerId:  w.ID,
						PendingAt: time.Now(),
					}
					d.store.AddPendingMsg(pendingMsg)
					msg.State = MSG_DEQUEUED
					d.store.PutMessage(msg)
					d.logger.Info("Pop message id:%v", string(msg.ID[:]))
				}
			}
			d.logger.Info("E: pop msg")
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
