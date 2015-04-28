package server

import (
	"sync"
)

type Dispatcher struct {
	topicName    string
	workers      map[string]*Worker
	workersMutex sync.RWMutex
	msgPopChan   chan *Message
	msgPushChan  chan *Message
	quitChan     chan struct{}
	running      bool
}

func NewDispatcher(topicName string) *Dispatcher {
	d := &Dispatcher{
		topicName:   topicName,
		workers:     make(map[string]*Worker),
		msgPopChan:  make(chan *Message),
		msgPushChan: make(chan *Message),
		quitChan:    make(chan struct{}),
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
			for _, w := range workers {
				ok := w.SendMessage(msg)
				if !ok {
					d.msgPushChan <- msg
				}
			}
		}
	}
}

func (d *Dispatcher) selectWorkers() []*Worker {
	var ws []*Worker
	for _, w := range d.workers {
		if w.Connected {
			ws = append(ws, w)
			break
		}
	}
	return ws
}
