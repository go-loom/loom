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
	if d.running == false {
		go d.dispatching()
		d.running = true
	}
}

func (d *Dispatcher) RemoveWorker(w *Worker) {
	d.workersMutex.Lock()
	defer d.workersMutex.Unlock()
	delete(d.workers, w.ID)
	if len(d.workers) <= 0 && d.running == true {
		d.quitChan <- struct{}{}
		d.running = false
	}
}

func (d *Dispatcher) dispatching() {
	for {
		select {
		case <-d.quitChan:
			break
		case msg := <-d.msgPopChan:
			workers := d.selectWorkers()
			if len(workers) <= 0 {
				d.msgPushChan <- msg
			}
			d.logger.Info("Dispatched Workers:%v", len(workers))
			for _, w := range workers {
				ok := w.SendMessage(msg)
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
		}
	}
}

func (d *Dispatcher) selectWorkers() []*Worker {
	var ws []*Worker
	for _, w := range d.workers {
		if w.Working() {
			ws = append(ws, w)
			break
		}
	}
	return ws
}
