package server

import (
	"errors"
	"path/filepath"
)

const (
	MessageBucketName                = "messages"
	MessagePendingBucketName         = "pendingMessages"
	WorkerPerMessageBucketNamePrefix = "worker:"
)

var ErrNotSupportedStoreType = errors.New("this store type is not supported")

type MessageBucket interface {
	Get(id MessageID) (*Message, error)
	Put(msg *Message) error
	Del(id MessageID) error
	Walk(walkFunc func(*Message) error) error
	DelBucket() error
}

type Store interface {
	Open() error
	Close() error
	MessageBucket(name string) MessageBucket
}

func NewTopicStore(storeType string, path string, topic string) (Store, error) {
	if storeType == "bolt" {
		path := filepath.Join(path, topic+".boltdb")
		store := NewBoltStore(path)
		return store, nil
	}
	return nil, ErrNotSupportedStoreType
}
