package server

import (
	"errors"
	"path/filepath"
	"time"
)

var ErrNotSupportedStoreType = errors.New("this store type is not supported")

type Store interface {
	GetMessage(id MessageID) (*Message, error)
	PutMessage(msg *Message) error
	RemoveMessage(id MessageID) error
	WalkMessage(walkFunc func(*Message) error) error

	AddPendingMsg(msg *PendingMessage) error
	RemovePendingMsgsInWorker(workerID string) error
	GetPendingMsgsInWorker(workerID string) ([]*PendingMessage, error)

	/*
		GetPendingMsgIdsBeforeTime(ts time.Time) ([]MessageID, error)
		UpdatePendingMsgIdsAtTime(ts time.Time, msgIds []MessageID) error
	*/

	SaveTasks(id MessageID, workerId string, tasks map[string]interface{}) error
	LoadTasks(id MessageID) (map[string]map[string]interface{}, error)

	Open() error
	Close() error
}

type PendingMessage struct {
	MessageID MessageID
	WorkerId  string
	PendingAt time.Time
}

func NewTopicStore(storeType string, path string, topic string) (Store, error) {
	if storeType == "bolt" {
		path := filepath.Join(path, topic+".boltdb")
		store := NewBoltStore(path)
		return store, nil
	}

	return nil, ErrNotSupportedStoreType
}
