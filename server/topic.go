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
	go topic.retryTick()

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
		t.push(m)

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

func (t *Topic) retryTick() {
	ticker := time.NewTicker(t.retryCheckDuration)
	for {
		select {
		case <-ticker.C:
			t.checkRetryJobs()
		case <-t.retryCheckQuitC:
			break
		}
	}

	t.logger.Info("Done retry checking")
}

func (t *Topic) checkRetryJobs() {
	t.pendingMsgBucket.Walk(func(m *Message) error {
		if m.Job.Retry != nil {
			retry := m.Job.Retry
			timeout, err := retry.GetTimeout()

			if err != nil {
				t.logger.Error("job.Retry timeout err:%v", err)
				return nil
			}

			if timeout != nil {
				if retry.NumRetry >= retry.Number {
					m.State = MSG_FAILURE
					err := t.msgBucket.Put(m)
					if err != nil {
						t.logger.Error("t.msgBucket.Put err:%v", err)
						return nil
					}
					err = t.pendingMsgBucket.Del(m.ID)
					if err != nil {
						t.logger.Error("t.pendingMsgBucket err:%v", err)
						return nil
					}
				} else if retry.NumRetry < retry.Number {

					checkedTime := retry.CheckedTime
					if retry.CheckedTime == nil {
						checkedTime = &m.Created
					}

					if time.Now().After(checkedTime.Add(*timeout)) {
						retry.IncrRetry()

						now := time.Now()
						retry.CheckedTime = &now

						m.State = MSG_PENDING
						t.msgBucket.Put(m)
						t.pendingMsgBucket.Put(m)
						if m.State == MSG_RECEIVED {
							t.push(m)
						}
						t.logger.Info("This message is timeout and is queueing again id:%s", m.ID[:])
					}
				}
			}
		}
		return nil
	})
}
