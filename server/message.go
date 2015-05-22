package server

import (
	"bytes"
	"encoding/gob"
	"gopkg.in/loom.v1/config"
	"time"
)

const MsgIdLen = 16

const (
	MSG_PENDING = iota
	MSG_RECEIVED
	MSG_FINISHED
	MSG_SUCCESS
	MSG_FAILURE
)

const (
	MsgPendingState  = "PENDING"
	MsgFinishedState = "FINISHED"
	MsgReceivedState = "RECEIVED"
	MsgSuccessState  = "SUCCESS"
	MsgFailureState  = "FAILURE"
)

var (
	MsgStates map[uint8]string
)

func init() {
	MsgStates = make(map[uint8]string)
	MsgStates[MSG_PENDING] = MsgPendingState
	MsgStates[MSG_RECEIVED] = MsgReceivedState
	MsgStates[MSG_SUCCESS] = MsgSuccessState
	MsgStates[MSG_FAILURE] = MsgFailureState

	gob.Register(map[string]interface{}{})
}

type MessageID [MsgIdLen]byte

type MessageIDGenerator interface {
	NewID() MessageID
}

type Message struct {
	ID      MessageID
	Job     *config.Job
	Created time.Time
	State   uint8
	Results *TaskResults
}

func (id MessageID) Bytes() []byte {
	return id[:]
}

func GetMessageID(v []byte) MessageID {
	var id MessageID
	copy(id[:], v)
	return id
}

func NewMessage(id MessageID, job *config.Job) *Message {
	m := &Message{
		ID:      id,
		Job:     job,
		Created: time.Now(),
		State:   MSG_PENDING,
	}

	return m
}

func DecodeMessage(v []byte) (*Message, error) {
	var msg Message
	buf := bytes.NewBuffer(v)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&msg)
	return &msg, err
}

func (m *Message) SetResults(workerId string, tasks map[string]interface{}) {
	if m.Results == nil {
		m.Results = NewTaskResults(workerId, tasks)
		return
	}
	m.Results.WorkerId = workerId
	m.Results.Tasks = tasks
	return
}

func (m *Message) Encode() []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(m)
	return buf.Bytes()
}

func (m *Message) JSON() Json {
	json := Json{
		"id":      string(m.ID[:]),
		"tasks":   m.Job.Tasks,
		"created": m.Created,
		"state":   MsgStates[m.State],
	}

	if m.Job.Retry != nil {
		json["retry"] = m.Job.Retry
	}

	return json
}
