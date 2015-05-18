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

func (w *Worker) SendMessage(msg *Message) bool {
	resp, err := w.Client.Tell("loom.worker:message.pop", msg.JSON())
	//TODO: Need better error handling
	if err != nil {
		return false
	}

	ok, okerr := resp.Bool()
	if err != nil || okerr != nil {
		return false
	}

	if ok {
		w.incrNumJob()
	}

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
	defer w.wMutex.Unlock()
	w.working = working
}

func (w *Worker) incrNumJob() {
	w.wMutex.Lock()
	defer w.wMutex.Unlock()

	w.numJob++
}

func (w *Worker) decrNumJob() {
	w.wMutex.Lock()
	defer w.wMutex.Unlock()

	w.numJob--
}

func (w *Worker) isNotOverMaxJobSize() bool {
	if w.numJob >= w.maxJobSize {
		return true
	}
	return false
}
