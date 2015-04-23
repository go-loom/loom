package server

import (
	"time"
)

const MsgIdLen = 16

const (
	MSG_ENQUEUED = iota
	MSG_DEQUEUED
	MSG_FINISHED
)

type MessageID [MsgIdLen]byte

type MessageIDGenerator interface {
	NewID() MessageID
}

type Message struct {
	ID      MessageID   `json:"id,string"`
	Value   interface{} `json:"value"`
	Created time.Time   `json:"created"`
	State   uint8       `json:"state"`
}

func NewMessage(id MessageID, value interface{}) *Message {
	m := &Message{
		ID:      id,
		Value:   value,
		Created: time.Now(),
		State:   MSG_ENQUEUED,
	}

	return m
}
