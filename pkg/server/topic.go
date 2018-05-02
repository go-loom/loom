package server

import (
	"github.com/go-loom/loom/pkg/log"

	kitlog "github.com/go-kit/kit/log"

	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type Topic struct {
	ctx                context.Context
	Name               string
	Queue              Queue
	retryCheckDuration time.Duration
	waitingCh          chan interface{}
	store              Store
	msgBucket          MessageBucket
	pendingMsgBucket   MessageBucket
	logger             kitlog.Logger
	quitC              chan struct{}
	retryCheckQuitC    chan struct{}
}

func NewTopic(ctx context.Context, name string, retryCheckDuration time.Duration, store Store) *Topic {

	topic := &Topic{
		ctx:                ctx,
		Name:               name,
		Queue:              NewLQueue(),
		retryCheckDuration: retryCheckDuration,
		waitingCh:          make(chan interface{}),
		store:              store,
		msgBucket:          store.MessageBucket(MessageBucketName),
		pendingMsgBucket:   store.MessageBucket(MessagePendingBucketName),
		logger:             log.With(log.Logger, "topic", name),
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

	return err
}

func (t *Topic) PushMessage(msg *Message) {
	msg.State = MSG_PENDING
	t.msgBucket.Put(msg)
	t.pendingMsgBucket.Put(msg)
	t.push(msg)

	log.Info(t.logger).Log("msg", "Pushed message", "id", string(msg.ID[:]))
}

func (t *Topic) FinishMessage(id MessageID) error {
	msg, err := t.msgBucket.Get(id)
	if err != nil {
		return err
	}

	msg.State = MSG_SUCCESS
	err = t.msgBucket.Put(msg)
	if err != nil {
		return err
	}

	err = t.pendingMsgBucket.Del(msg.ID)
	if err != nil {
		return err
	}

	if msg.Job.FinishReportURL != "" {
		url := msg.Job.FinishReportURL
		msgJson, err := json.Marshal(msg.JSON())
		if err != nil {
			log.Error(t.logger).Log("msg", "FinishReportURL json", "err", err)
			return err
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(msgJson))
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Error(t.logger).Log("msg", "FinishReportURL request", "err", err)
			return err
		}
		defer resp.Body.Close()
	}

	log.Info(t.logger).Log("msg", "Finished message", "id", string(msg.ID.Bytes()))
	return nil
}

/*
func (t *Topic) PushPendingMsgsInWorker(workerId string) {
	pmb := t.store.MessageBucket(WorkerPerMessageBucketNamePrefix + workerId)

	num := 0
	pmb.Walk(func(pm *Message) error {
		m, err := t.msgBucket.Get(pm.ID)
		if err != nil {
			return err
		}
		if m.State == MSG_RECEIVED {
			m.State = MSG_PENDING
			err = t.msgBucket.Put(m)
			if err != nil {
				t.logger.Error("Error message %v requeue", string(m.ID[:]))
			}
			t.push(m)

			t.logger.Info("Message %v is re-pushed to queue", string(m.ID[:]))
			num++
		}
		return err
	})

	err := pmb.DelBucket()
	if err != nil {
		t.logger.Error("PushPendingMsgsInWorker Remove err:%v", err)
	}
	t.logger.Info("Pushed pending msgs:%v", num)
	return
}
*/

func (t *Topic) push(msg *Message) {
	t.Queue.Push(msg)
}

func (t *Topic) PopMessage() *Message {
	return t.pop()
}

func (t *Topic) pop() (msg *Message) {
	if item := t.Queue.Pop(); item != nil {
		msg = item.(*Message)
	}
	return
}

func (t *Topic) waitDone() {
	<-t.ctx.Done()
	close(t.waitingCh)
	t.store.Close()
	t.retryCheckQuitC <- struct{}{}
	t.quitC <- struct{}{}

	log.Info(t.logger).Log("msg", "Closing topic")
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

	log.Info(t.logger).Log("msg", "Done retry checking...")
}

func (t *Topic) checkRetryJobs() {
	t.pendingMsgBucket.Walk(func(m *Message) error {
		if m.Job.Retry != nil {
			retry := m.Job.Retry
			timeout, err := retry.GetTimeout()

			if err != nil {
				log.Error(t.logger).Log("msg", "job.Retry timeout", "err", err)
				return nil
			}

			if timeout != nil {
				if retry.NumRetry > retry.Number {
					m.State = MSG_FAILURE
					err := t.msgBucket.Put(m)
					if err != nil {
						log.Error(t.logger).Log("msg", "t.msgBucket.Put", "err", err)
						return nil
					}
					err = t.pendingMsgBucket.Del(m.ID)
					if err != nil {
						log.Error(t.logger).Log("msg", "t.pendingMsgBucket", "err", err)
						return nil
					}
					log.Error(t.logger).Log("msg", "Taken maxretry count", "id", string(m.ID[:]), "num", retry.Number)

				} else if retry.NumRetry < retry.Number {

					checkedTime := retry.CheckedTime
					if retry.CheckedTime == nil {
						checkedTime = &m.Created
					}

					if time.Now().After(checkedTime.Add(*timeout)) {
						retry.IncrRetry()

						now := time.Now()
						retry.CheckedTime = &now

						if m.State == MSG_RECEIVED {
							m.State = MSG_PENDING
						}
						t.msgBucket.Put(m)
						t.pendingMsgBucket.Put(m)
						t.push(m)

						log.Info(t.logger).Log("msg", "This message is timeout and queueing againg", "id", m.ID[:])
					}
				}
			}
		}
		return nil
	})
}
