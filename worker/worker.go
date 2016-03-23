package worker

import (
	"fmt"
	"github.com/koding/kite"
	"golang.org/x/net/context"
	"gopkg.in/loom.v1/config"
	"gopkg.in/loom.v1/log"
	"sync"
	"time"
)

type Worker struct {
	ID         string
	Topic      string
	ServerURL  string
	maxJobSize int
	kite       *kite.Kite
	Client     *kite.Client
	jobs       map[string]*Job
	jobsMutex  sync.RWMutex
	logger     log.Logger
	ctx        context.Context
}

func NewWorker(serverURL, topic string, maxJobSize int, k *kite.Kite) *Worker {
	w := &Worker{
		ID:         k.Id,
		Topic:      topic,
		ServerURL:  serverURL,
		maxJobSize: maxJobSize,
		jobs:       make(map[string]*Job, 0),
		kite:       k,
		logger:     log.New("worker#" + topic),
		ctx:        context.Background(), //TODO:
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

func (w *Worker) Run() {
	w.kite.Run()
}

func (w *Worker) Close() error {
	w.logger.Info("Closing worker...")
	w.tellWorkerShutdown()
	w.kite.Close()
	return nil
}

func (w *Worker) tellHelloToServer() error {
	response, err := w.Client.Tell("loom.server:worker.connect", w.ID, w.Topic, w.maxJobSize)
	if err != nil {
		return err
	}

	ok := response.MustBool()
	if !ok {
		return fmt.Errorf("registering worker to server is failed")
	}

	return nil
}

func (w *Worker) tellWorkerInfo() {
	numJob := 0

	w.jobsMutex.RLock()
	numJob = len(w.jobs)
	w.jobsMutex.RUnlock()

	_, err := w.Client.Tell("loom.server:worker.info", numJob)
	if err != nil {
		w.logger.Error("tellWorkerInfo err:%v", err)
	}

	w.logger.Info("Running job:%v (tell worker info to server)", numJob)
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
		w.logger.Error("can't tell JobDone to server err: %v", err)
		return err
	}
	return nil
}

func (w *Worker) tellWorkerShutdown() error {
	return nil

	/*
		w.logger.Info("tellWorkerShutdown:1")
		//TODO: Retry?
		_, err := w.Client.Tell("loom.server:worker.shutdown")

		w.logger.Info("tellWorkerShutdown:2")
		if err != nil {
			w.logger.Error("can't tell worker shutdown err:%v", err)
			return err
		}

		w.logger.Info("tellWorkerShutdown:3")
		return nil
	*/
}

func (w *Worker) HandleMessagePop(r *kite.Request) (interface{}, error) {
	msg, err := r.Args.One().Map()
	if err != nil {
		return false, err
	}

	//Consider this job is working or new.
	w.jobsMutex.RLock()
	jobId := msg["id"].MustString()
	if _, ok := w.jobs[jobId]; ok {
		w.logger.Info("Job has already working.")
		return true, nil
	}
	w.jobsMutex.RUnlock()

	var tasksConfig *[]*config.Task
	if err := msg["tasks"].Unmarshal(&tasksConfig); err != nil {
		w.logger.Error("json err: %v", err)
		return false, err
	}

	w.logger.Info("tasksConfig:%v", tasksConfig)

	jobConfig := &config.Job{
		Tasks: *tasksConfig,
	}

	job := NewJob(w.ctx, msg["id"].MustString(), jobConfig)
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

		w.tellWorkerInfo()

	}()
	job.Run() //async.

	w.jobsMutex.Lock()
	w.jobs[job.ID] = job
	w.jobsMutex.Unlock()

	go func() {
		w.tellWorkerInfo()
	}()

	w.logger.Info("Recieved job id:%v", job.ID)

	return true, nil
}
