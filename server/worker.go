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
