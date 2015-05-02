package server

import (
	"github.com/koding/kite"
)

//Worker represents connected worker
type Worker struct {
	ID        string
	TopicName string
	Client    *kite.Client
	Connected bool
}

func NewConnectedWorker(id, topicName string, client *kite.Client) *Worker {
	w := &Worker{
		ID:        id,
		TopicName: topicName,
		Client:    client,
		Connected: true,
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
