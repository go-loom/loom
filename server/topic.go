package server

import (
	"log"
	"sync"
	"time"
)

type Topic struct {
	Name                 string
	Queue                Queue
	Dispatcher           *Dispatcher
	PendingTimeout       time.Duration
	pendingCheckInterval time.Duration
	waitingCh            chan interface{}
	mutex                sync.Mutex
	store                Store
}

func NewTopic(name string, pendingTimeout time.Duration, store Store) *Topic {
	dispatcher := NewDispatcher(name)

	topic := &Topic{
		Name:                 name,
		Queue:                NewLLQueue(),
		Dispatcher:           dispatcher,
		PendingTimeout:       pendingTimeout,
		pendingCheckInterval: 10 * time.Second,
		waitingCh:            make(chan interface{}),
		store:                store,
	}

	return topic
}

func (t *Topic) Init() error {

	//First time, Messages go from Disk to Queue.
	err := t.store.WalkMessage(func(m *Message) error {
		if m.State == MSG_ENQUEUED {
			t.push(m)
		}
		return nil
	})

	go t.msgDispatch()
	go t.pendingChecker()

	return err
}

func (t *Topic) PushMessage(msg *Message) {
	t.store.PutMessage(msg)
	t.push(msg)
}

func (t *Topic) FinishMessage(id MessageID) error {
	//TODO:
	/*
		msg, err := t.store.GetMessage(id)
		if err != nil {
			return err
		}

			if msg.State == MSG_ENQUEUED {
				return fmt.Errorf("Message complete failed")
			}
	*/

	err := t.store.RemoveMessage(id)
	return err
}

func (t *Topic) push(msg *Message) {
	select {
	case t.waitingCh <- msg:
	default:
		t.mutex.Lock()
		defer t.mutex.Unlock()
		err := t.Queue.Push(msg)
		if err != nil {
			//TODO:
			log.Println(err)
		}
	}
}

func (t *Topic) pop() (msg *Message) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if item := t.Queue.Pop(); item != nil {
		msg = item.(*Message)
		msg.State = MSG_DEQUEUED
	} else {
		item := <-t.waitingCh
		msg = item.(*Message)
		msg.State = MSG_DEQUEUED
	}

	return

}

func (t *Topic) msgDispatch() {
	for {
		select {
		case msg := <-t.Dispatcher.msgPushChan:
			t.PushMessage(msg)
		default:
			msg := t.pop()
			t.Dispatcher.msgPopChan <- msg
			now := time.Now()
			t.store.PutMessage(msg)
			t.store.PutPendingMsgID(&now, msg.ID)
		}
	}
}

func (t *Topic) pendingChecker() {

	//st := time.Now()
	st, _ := time.Parse("2006-Jan-02", "2013-Feb-03")

	walkFunc := func(ts *time.Time, id MessageID) error {
		if ts.Add(t.PendingTimeout).Before(time.Now()) {
			m, err := t.store.GetMessage(id)
			if err != nil {
				//TODO: Log
			}

			m.State = MSG_ENQUEUED
			t.store.PutMessage(m)
			t.store.RemovePendingMsgID(ts)
			t.push(m)
		}

		return nil
	}

	ticker := time.NewTicker(t.pendingCheckInterval)
	for {
		select {
		case <-ticker.C:
			ed := st.Add(t.pendingCheckInterval)
			err := t.store.WalkPendingMsgId(&st, &ed, walkFunc)
			if err != nil {
				//TODO: Log

			}
			st = ed
		}
	}

}
