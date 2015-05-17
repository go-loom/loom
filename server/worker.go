package server

import (
	"github.com/koding/kite"
	"sync"
)

//Worker represents connected worker
type Worker struct {
	ID        string
	TopicName string
	Client    *kite.Client
	working   bool
	wMutex    sync.RWMutex
}

func NewConnectedWorker(id, topicName string, client *kite.Client) *Worker {
	w := &Worker{
		ID:        id,
		TopicName: topicName,
		Client:    client,
		working:   true,
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
	return ok
}

func (w *Worker) Working() bool {
	w.wMutex.RLock()
	defer w.wMutex.RUnlock()
	return w.working
}

func (w *Worker) SetWorking(working bool) {
	w.wMutex.Lock()
	defer w.wMutex.Unlock()
	w.working = working
}
