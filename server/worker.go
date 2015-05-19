package server

import (
	"github.com/koding/kite"
	"sync"
)

//Worker represents connected worker
type Worker struct {
	ID         string
	TopicName  string
	Client     *kite.Client
	dispatcher *Dispatcher
	maxJobSize int
	numJob     int
	working    bool
	wMutex     sync.RWMutex
}

func NewConnectedWorker(id, topicName string, maxJobSize int, client *kite.Client) *Worker {
	w := &Worker{
		ID:         id,
		TopicName:  topicName,
		Client:     client,
		maxJobSize: maxJobSize,
		working:    true,
	}
	return w
}

func (w *Worker) SetDispatcher(dispatcher *Dispatcher) {
	w.dispatcher = dispatcher
}

func (w *Worker) SendMessage(msg *Message) bool {
	logger.Info("Sent from worker")
	resp, err := w.Client.Tell("loom.worker:message.pop", msg.JSON())
	logger.Info("Sent from worker")
	//TODO: Need better error handling
	if err != nil {
		return false
	}

	ok, okerr := resp.Bool()
	if err != nil || okerr != nil {
		return false
	}

	logger.Info("Sent from worker")
	return ok
}

func (w *Worker) Working() bool {
	w.wMutex.RLock()
	defer w.wMutex.RUnlock()
	if w.working && w.isNotOverMaxJobSize() {
		return true
	}
	return false
}

func (w *Worker) SetWorking(working bool) {
	w.wMutex.Lock()
	w.working = working
	w.wMutex.Unlock()

	if w.dispatcher != nil {
		w.dispatcher.NotifyWorkerState()
	}
}

func (w *Worker) SetNumJob(numJob int) {
	w.wMutex.Lock()
	w.numJob = numJob
	w.wMutex.Unlock()

	if w.dispatcher != nil {
		w.dispatcher.NotifyWorkerState()
	}
}

func (w *Worker) isNotOverMaxJobSize() bool {
	if w.numJob < w.maxJobSize {
		return true
	}
	return false
}
