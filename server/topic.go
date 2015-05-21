package server

import (
	"golang.org/x/net/context"
	"gopkg.in/loom.v1/log"
	"time"
)

type Topic struct {
	ctx                context.Context
	Name               string
	Queue              Queue
	Dispatcher         *Dispatcher
	retryCheckDuration time.Duration
	waitingCh          chan interface{}
	store              Store
	msgBucket          MessageBucket
	pendingMsgBucket   MessageBucket
	logger             log.Logger
	quitC              chan struct{}
	retryCheckQuitC    chan struct{}
}

func NewTopic(ctx context.Context, name string, retryCheckDuration time.Duration, store Store) *Topic {
	dispatcher := NewDispatcher(name, store)

	topic := &Topic{
		ctx:                ctx,
		Name:               name,
		Queue:              NewLQueue(),
		Dispatcher:         dispatcher,
		retryCheckDuration: retryCheckDuration,
		waitingCh:          make(chan interface{}),
		store:              store,
		msgBucket:          store.MessageBucket(MessageBucketName),
		pendingMsgBucket:   store.MessageBucket(MessagePendingBucketName),
		logger:             log.New("Topic:" + name),
		quitC:              ctx.Value("quitC").(chan struct{}),
		retryCheckQuitC:    make(chan struct{}),
	}

	go topic.waitDone()
	go topic.retryChecking()

	return topic
}

func (t *Topic) Init() error {

	//First time, Messages go from Disk to Queue.
	err := t.msgBucket.Walk(func(m *Message) error {
		if m.State == MSG_PENDING {
			t.push(m)
		}
		return nil
	})

	go t.msgPopDispatch()
	go t.msgPushDispatch()

	return err
}

func (t *Topic) PushMessage(msg *Message) {
	msg.State = MSG_PENDING
	t.msgBucket.Put(msg)
	t.pendingMsgBucket.Put(msg)
	t.push(msg)

	t.logger.Info("Pushed message id:%s", string(msg.ID[:]))
}

func (t *Topic) FinishMessage(id MessageID) error {
	msg, err := t.msgBucket.Get(id)
	if err != nil {
		return err
	}

	msg.State = MSG_FINISHED
	err = t.msgBucket.Put(msg)
	if err != nil {
		return err
	}

	err = t.pendingMsgBucket.Del(msg.ID)
	if err != nil {
		return err
	}

	t.logger.Info("Finished message id:%v", string(msg.ID.Bytes()))
	return nil
}

func (t *Topic) PushPendingMsgsInWorker(workerId string) {
	pmb := t.store.MessageBucket(WorkerPerMessageBucketNamePrefix + workerId)

	num := 0
	pmb.Walk(func(pm *Message) error {
		m, err := t.msgBucket.Get(pm.ID)
		if err != nil {
			return err
		}
		m.State = MSG_PENDING
		err = t.msgBucket.Put(m)
		if err != nil {
			t.logger.Error("Error message %v requeue", string(m.ID[:]))
		}

		t.logger.Info("Message %v is re-pushed to queue", string(m.ID[:]))
		num++
		return err
	})

	err := pmb.DelBucket()
	if err != nil {
		t.logger.Error(" PushPendingMsgsInWorker Remove err:%v", err)
	}
	t.logger.Info("Pushed pending msgs:%v", num)
	return
}

func (t *Topic) push(msg *Message) {
	select {
	case t.waitingCh <- msg:
	default:
		t.Queue.Push(msg)
	}
}

func (t *Topic) pop() (msg *Message) {

	if item := t.Queue.Pop(); item != nil {
		msg = item.(*Message)
	} else {
		item, ok := <-t.waitingCh
		if !ok {
			return nil
		}
		msg = item.(*Message)
	}

	return

}

func (t *Topic) msgPopDispatch() {
	for {
		msg := t.pop()
		if msg == nil {
			break
		}
		t.logger.Debug("S:Pop job from queue id:%v", string(msg.ID[:]))
		t.Dispatcher.msgPopChan <- msg
		t.logger.Debug("E:Pop job from queue id:%v", string(msg.ID[:]))
	}
}

func (t *Topic) msgPushDispatch() {
	for {
		select {
		case msg := <-t.Dispatcher.msgPushChan:
			t.logger.Debug("dispatcher.push")
			t.PushMessage(msg)
			if t.logger.IsDebug() {
				t.logger.Debug("From dispatcher msg push id:%v", string(msg.ID[:]))
			}
		}
	}
}

func (t *Topic) waitDone() {
	<-t.ctx.Done()
	close(t.waitingCh)
	t.store.Close()
	t.retryCheckQuitC <- struct{}{}
	t.quitC <- struct{}{}

	t.logger.Info("Closing topic")
}

func (t *Topic) retryChecking() {

	ticker := time.NewTicker(t.retryCheckDuration)
	for {
		select {
		case <-ticker.C:
			t.logger.Info("Check retry messages")
		case <-t.retryCheckQuitC:
			break
		}
	}

	t.logger.Info("Done retry checking")
}

/*
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
*/
