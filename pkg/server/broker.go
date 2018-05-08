package server

import (
	"github.com/go-loom/loom/pkg/config"
	"github.com/go-loom/loom/pkg/log"
	"github.com/go-loom/loom/pkg/rpc/pb"

	kitlog "github.com/go-kit/kit/log"

	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	ErrTopicNotFound = errors.New("Topic not found")
	ErrMsgNotFound   = errors.New("Message not found")
)

type Broker struct {
	ctx        context.Context
	ID         int64
	DBPath     string
	Topics     map[string]*Topic
	topicMutex sync.Mutex

	action chan func()

	idChan     chan MessageID
	topicQuitC chan struct{}

	quitC chan struct{}
	wg    sync.WaitGroup

	logger kitlog.Logger
}

func NewBroker(ctx context.Context, dbpath string) *Broker {

	b := &Broker{
		ctx:        ctx,
		ID:         1,
		DBPath:     dbpath,
		Topics:     make(map[string]*Topic),
		action:     make(chan func()),
		idChan:     make(chan MessageID, 4096), // Buffer
		topicQuitC: make(chan struct{}),
		quitC:      make(chan struct{}),
		logger:     log.With(log.Logger),
	}

	go b.loop()
	go b.recvTopicQuit()
	return b
}

func (b *Broker) Init() error {
	go b.idPump()

	if _, err := os.Stat(b.DBPath); os.IsNotExist(err) {
		err := os.MkdirAll(b.DBPath, 0777)
		if err != nil {
			return err
		}
	}

	pattern := filepath.Join(b.DBPath, "*.boltdb")
	dbfilepaths, err := filepath.Glob(pattern)
	if err != nil {
		log.Error(b.logger).Log("msg", "Wrong dbpath", "err", err)
		return err
	}

	for _, p := range dbfilepaths {
		topicName := strings.Replace(filepath.Base(p), filepath.Ext(p), "", 1)
		b.Topic(topicName)
		log.Info(b.logger).Log("msg", "load topic db ", "topic", topicName, "db", p)
	}

	return nil
}

func (b *Broker) Done() {
	b.wg.Wait()
	log.Info(b.logger).Log("msg", "closing broker...")
	return
}

// twirp rpc interface
func (b *Broker) SubscribeJob(ctx context.Context, req *pb.SubscribeJobRequest) (*pb.SubscribeJobResponse, error) {

	var (
		notFound  = make(chan struct{})
		otherErr  = make(chan error)
		newJobMsg = make(chan *Message)
	)

	topicName := req.TopicName
	workerID := req.WorkerId

	l := log.With(b.logger, "f", "SubscribeJob", "worker", workerID, "topic", topicName)
	log.Debug(l).Log("start", true)

	b.action <- func() {
		topic := b.Topic(topicName)
		if topic == nil {
			log.Error(l).Log("err", ErrTopicNotFound)
			otherErr <- ErrTopicNotFound
			return
		}
		msg := topic.PopMessage()
		if msg == nil {
			notFound <- struct{}{}
			return
		}
		newJobMsg <- msg
	}

	var (
		err error
		res *pb.SubscribeJobResponse = &pb.SubscribeJobResponse{}
	)

	select {
	case <-notFound:
		res.JobStatus = pb.SubscribeJobResponse_NoJob
	case err = <-otherErr:
	case msg := <-newJobMsg:
		res.JobStatus = pb.SubscribeJobResponse_NewJob
		res.JobId = msg.ID.Bytes()
		b, _err := json.Marshal(msg.JSON())
		if _err != nil {
			err = _err
			log.Error(l).Log("err", err)
			return res, err
		}
		res.JobMsg = b
		log.Info(l).Log("newJob", res.JobId)
	}

	log.Debug(l).Log("end", true)
	return res, err
}

func (b *Broker) ReportJob(ctx context.Context, req *pb.ReportJobRequest) (res *pb.ReportJobResponse, err error) {
	jobID := req.JobId
	workerID := req.WorkerId
	topicName := req.TopicName
	jobMsg := req.JobMsg

	l := log.With(b.logger, "f", "ReportJob", "worker", workerID, "topic", topicName, "job", jobID)

	topic := b.Topic(topicName)
	msgID := GetMessageID(jobID)
	msg, err := topic.msgBucket.Get(msgID)
	msg.State = MSG_RECEIVED
	if err != nil {
		log.Error(l).Log("err", err)
		return
	}
	if msg == nil {
		log.Error(l).Log("err", ErrMsgNotFound)
	}

	var tasks *map[string]interface{}
	err = json.Unmarshal(jobMsg, tasks)
	if err != nil {
		log.Error(l).Log("err", err)
	}

	msg.SetResults(workerID, *tasks)
	err = topic.msgBucket.Put(msg)
	if err != nil {
		log.Error(l).Log("err", err)
	}

	log.Info(l).Log("msg", "Received task results")

	return
}

func (b *Broker) ReportJobDone(ctx context.Context, req *pb.ReportJobDoneRequest) (res *pb.ReportJobDoneResponse, err error) {
	jobID := req.JobId
	workerID := req.WorkerId
	topicName := req.TopicName
	l := log.With(b.logger, "f", "ReportJobDone", "worker", workerID, "topic", topicName, "job", string(jobID))

	topic := b.Topic(topicName)

	err = topic.FinishMessage(GetMessageID(jobID))
	if err != nil {
		log.Error(l).Log("err", err)
		return
	}

	l.Log()
	return
}

func (b *Broker) Topic(name string) *Topic {
	b.topicMutex.Lock()
	defer b.topicMutex.Unlock()

	if t, ok := b.Topics[name]; ok {
		return t
	}

	topicCtx := context.WithValue(b.ctx, "quitC", b.topicQuitC)

	store, _ := NewTopicStore("bolt", b.DBPath, name)
	store.Open()

	pendingTimeout := 10 * time.Second
	t := NewTopic(topicCtx, name, pendingTimeout, store)
	err := t.Init()
	if err != nil {
		log.Error(b.logger).Log("msg", "load topic", "topic", name, "err", err)
	}

	b.Topics[name] = t

	b.wg.Add(1) //Add topic wait group count

	return t
}

func (b *Broker) PushMessage(name string, job *config.Job) (*Message, error) {
	t := b.Topic(name)
	msg := NewMessage(b.NewID(), job)
	t.PushMessage(msg)
	return msg, nil
}

func (b *Broker) GetMessage(name string, id MessageID) (*Message, error) {
	t := b.Topic(name)

	msg, err := t.msgBucket.Get(id)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (b *Broker) FinishMessage(name string, id MessageID) error {
	t := b.Topic(name)
	err := t.FinishMessage(id)
	return err
}

//This method is implemented to MessageIDGenerator
func (b *Broker) NewID() MessageID {
	return <-b.idChan
}

func (b *Broker) recvTopicQuit() {
	for _ = range b.topicQuitC {
		b.wg.Done()
	}
}

func (b *Broker) loop() {
	for {
		select {
		case f := <-b.action:
			f()
		case <-b.ctx.Done():
			return
		}
	}
}

func (b *Broker) idPump() {
	factory := &guidFactory{}
	lastError := time.Now()

	for {

		id, err := factory.NewGUID(b.ID)

		if err != nil {
			now := time.Now()
			if now.Sub(lastError) > time.Second {
				//only print the error once/second
				//TODO:
				lastError = now
				log.Error(b.logger).Log("msg", "id pump", "err", err)

			}
			runtime.Gosched()
			continue
		}

		select {
		case b.idChan <- id.Hex():
			//TODO: exitChan?
		case <-b.ctx.Done():
			break
		}
	}

	log.Debug(b.logger).Log("msg", "end idpump")
}
