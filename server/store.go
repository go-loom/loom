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

	//GetPendingMsgIDList(st *time.Time, et *time.Time) ([]MessageID, error)
	WalkPendingMsgId(st *time.Time, et *time.Time,
		walkFunc func(ts *time.Time, id MessageID) error) error
	PutPendingMsgID(ts *time.Time, id MessageID) error
	RemovePendingMsgID(ts *time.Time) error

	SaveTasks(id MessageID, workerId string, tasks []map[string]interface{}) error
	LoadTasks(id MessageID) (map[string][]map[string]interface{}, error)

	Open() error
	Close() error
}

func NewTopicStore(storeType string, path string, topic string) (Store, error) {
	if storeType == "bolt" {
		path := filepath.Join(path, topic+".boltdb")
		store := NewBlotStore(path)
		return store, nil
	}

	return nil, ErrNotSupportedStoreType
}
