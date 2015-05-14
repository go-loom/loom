package worker

import (
	"encoding/json"
	"fmt"
	"github.com/koding/kite"
	"golang.org/x/net/context"
	"gopkg.in/loom.v1/log"
	"gopkg.in/loom.v1/worker/config"
	"sync"
	"time"
)

type Worker struct {
	ID        string
	Topic     string
	ServerURL string
	kite      *kite.Kite
	Client    *kite.Client
	jobs      map[string]*Job
	jobsMutex sync.RWMutex
	logger    log.Logger
	ctx       context.Context
}

func NewWorker(serverURL, topic string, k *kite.Kite) *Worker {
	w := &Worker{
		ID:        k.Id,
		Topic:     topic,
		ServerURL: serverURL,
		jobs:      make(map[string]*Job, 0),
		kite:      k,
		logger:    log.New("worker-" + topic),
		ctx:       context.Background(), //TODO:
	}

	return w
}

func (w *Worker) Init() error {
	w.Client = w.kite.NewClient(w.ServerURL)

	w.logger.Info("Worker ID: %v", w.ID)

	connected, err := w.Client.DialForever()
	<-connected
	if err != nil {
		return err
	}

	w.Client.OnConnect(func() {
		err := w.tellHelloToServer()
		if err != nil {
			return
		}
	})

	w.tellHelloToServer()
	w.kite.HandleFunc("loom.worker:message.pop", w.HandleMessagePop)

	return nil
}

func (w *Worker) tellHelloToServer() error {
	response, err := w.Client.Tell("loom.server:worker.connect", w.ID, w.Topic)
	if err != nil {
		return err
	}

	ok := response.MustBool()
	if !ok {
		return fmt.Errorf("registering worker to server is failed")
	}

	return nil
}

func (w *Worker) tellJobTaskStateChange(job *Job, tasks Tasks) error {
	_, err := w.Client.Tell("loom.server:worker.job.tasks.state",
		tasks.JSON(), job.ID, w.Topic)
	if err != nil {
		w.logger.Error("tellJobTaskStateChange err:%v", err)
		return err
	}

	return nil
}

func (w *Worker) tellJobDone(job *Job) error {
	//TODO: Retry
	_, err := w.Client.Tell("loom.server:worker.job.done", job.ID, w.Topic)
	if err != nil {
		w.logger.Error("tellJobDoneToServer err: %v", err)
		return err
	}
	return nil
}

func (w *Worker) HandleMessagePop(r *kite.Request) (interface{}, error) {
	msg, err := r.Args.One().Map()
	if err != nil {
		return false, err
	}

	value := msg["value"].MustString()
	var jobConfig config.Job
	err = json.Unmarshal([]byte(value), &jobConfig)
	if err != nil {
		w.logger.Error("json err: %v", err)
		return false, err
	}

	if w.logger.IsDebug() {
		w.logger.Debug("message id: %v value: %v", msg["id"], jobConfig)
	}

	job := NewJob(w.ctx, msg["id"].MustString(), &jobConfig)

	job.OnTaskStateChange(func(task Task) {
		tasks := make(Tasks)

		for k, v := range job.Tasks {
			if k == task.TaskName() {
				tasks[k] = task
			} else {
				tasks[k] = v
			}
		}

		w.tellJobTaskStateChange(job, tasks)
	})

	//When job done.
	go func() {
		<-job.Done()
		for err := w.tellJobDone(job); err != nil; {
			time.Sleep(1 * time.Second)
		}

		w.jobsMutex.Lock()
		delete(w.jobs, job.ID)
		w.jobsMutex.Unlock()

	}()
	job.Run() //async.

	w.jobsMutex.Lock()
	w.jobs[job.ID] = job
	w.jobsMutex.Unlock()

	if w.logger.IsDebug() {
		w.logger.Debug("Pop message id: %v job: %v", job.ID, job)
	}

	return true, nil
}
