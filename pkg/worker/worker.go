package worker

import (
	"github.com/go-loom/loom/pkg/config"
	"github.com/go-loom/loom/pkg/log"
	"github.com/go-loom/loom/pkg/rpc/pb"

	kitlog "github.com/go-kit/kit/log"

	"context"
	"encoding/json"
	"sync"
	"time"
)

type Worker struct {
	Name       string
	Topic      string
	ServerURL  string
	maxJobSize int
	client     *Client
	jobq       chan *Job
	jobs       map[string]*Job
	jobsMutex  sync.RWMutex
	logger     kitlog.Logger
	ctx        context.Context
	stop       chan chan struct{}
}

func NewWorker(ctx context.Context, name string, serverURL, topic string, maxJobSize int) *Worker {
	client := NewClient(serverURL)
	w := &Worker{
		Name:       name,
		Topic:      topic,
		ServerURL:  serverURL,
		maxJobSize: maxJobSize,
		client:     client,
		jobs:       make(map[string]*Job, 0),
		logger:     log.With(log.Logger, "worker", topic),
		ctx:        ctx,
		jobq:       make(chan *Job, maxJobSize),
		stop:       make(chan chan struct{}),
	}

	for i := 0; i < w.maxJobSize; i++ {
		go w.processJob()
	}

	go w.loop()

	return w
}

func (w *Worker) Init() error {
	return nil
}

func (w *Worker) Run() {
	<-w.ctx.Done()
	w.Stop()
}

func (w *Worker) Stop() {
	c := make(chan struct{})
	w.stop <- c
	<-c
}

func (w *Worker) loop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			req := &pb.SubscribeJobRequest{
				WorkerId:  w.Name,
				TopicName: w.Topic,
			}
			res, err := w.client.SubscribeJob(w.ctx, req)
			if err != nil {
				log.Error(w.logger).Log("msg", "subscribe job", "err", err)
				time.Sleep(1 * time.Second)
			}

			if res.JobStatus == pb.SubscribeJobResponse_NoJob {
				time.Sleep(1 * time.Second)
				continue
			}

			{
				if w.isWorkingJob(string(res.JobId)) {
					log.Info(w.logger).Log("msg", "the job has already worked")
					continue
				}
				job, err := w.newJob(res)
				if err != nil {
					log.Error(w.logger).Log("msg", "map job", "err", err)
					continue
				}
				w.jobq <- job
			}

		case c := <-w.stop:
			close(c)
			return
		}
	}
}

func (w *Worker) isWorkingJob(jobID string) bool {
	w.jobsMutex.RLock()
	defer w.jobsMutex.RUnlock()
	if _, ok := w.jobs[jobID]; ok {
		return true
	}
	return false
}

func (w *Worker) newJob(res *pb.SubscribeJobResponse) (*Job, error) {
	var tasksConfig *[]*config.Task

	err := json.Unmarshal(res.JobMsg, tasksConfig)
	if err != nil {
		log.Error(w.logger).Log("msg", "jobmsg", "err", err)
		//TODO: client Report job (error)
		return nil, err
	}

	//log.Debug(w.logger).Log("tasksConfig", tasksConfig)

	jobConfig := &config.Job{
		Tasks: *tasksConfig,
	}

	jobID := string(res.JobId)

	job := NewJob(w.ctx, jobID, jobConfig)
	job.OnTaskStateChange(func(task Task) {
		tasks := make(Tasks)

		for k, v := range job.Tasks {
			if k == task.TaskName() {
				tasks[k] = task
			} else {
				tasks[k] = v
			}
		}

		w.reportJob(job, tasks)
	})

	return job, nil
}

func (w *Worker) processJob() {
	job := <-w.jobq
	go func() {
		<-job.Done()
		for err := w.reportJobDone(job); err != nil; {
			log.Error(w.logger).Log("err", err)
			time.Sleep(1 * time.Second)
		}

		w.jobsMutex.Lock()
		delete(w.jobs, job.ID)
		w.jobsMutex.Unlock()

	}()
	job.Run() // async

}

func (w *Worker) reportJobDone(job *Job) error {
	req := &pb.ReportJobDoneRequest{
		JobId:     []byte(job.ID),
		TopicName: w.Topic,
		WorkerId:  w.Name,
	}
	_, err := w.client.ReportJobDone(w.ctx, req)
	if err != nil {
		log.Error(w.logger).Log("job", job.ID, "err", err)
		return err
	}

	return nil
}

func (w *Worker) reportJob(job *Job, tasks Tasks) error {
	msg, err := json.Marshal(tasks.JSON())
	if err != nil {
		log.Error(w.logger).Log("job", job.ID, "err", err)
		return err
	}
	req := &pb.ReportJobRequest{
		JobId:     []byte(job.ID),
		TopicName: w.Topic,
		WorkerId:  w.Name,
		JobMsg:    msg,
	}
	if _, err = w.client.ReportJob(w.ctx, req); err != nil {
		log.Error(w.logger).Log("job", job.ID, "err", err)
		return err
	}
	return nil
}
