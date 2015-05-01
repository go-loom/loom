package worker

import (
	"github.com/mgutz/logxi/v1"
	"golang.org/x/net/context"
	"gopkg.in/loom.v1/worker/config"
)

type WorkFunc func() WorkFunc

type Work struct {
	ctx         context.Context
	cancelFunc  context.CancelFunc
	Job         Job
	taskConfigs []*config.Task
	tasks       Tasks
	running     bool
	logger      log.Logger
}

func NewWork(ctx context.Context, job Job, tasks []*config.Task) *Work {
	_ctx, cancelFunc := context.WithCancel(ctx)
	w := &Work{
		ctx:         _ctx,
		cancelFunc:  cancelFunc,
		Job:         job,
		taskConfigs: tasks,
		logger:      log.New("work#" + job["ID"].(string)),
	}

	return w

}

func (w *Work) Cancel() {
	w.cancelFunc()
}

func (w *Work) Done() <-chan struct{} {
	return w.ctx.Done()
}

func (w *Work) Run() {
	go func() {
		w.logger.Info("start")
		w.running = true
		for workFunc := w.workStartJob; workFunc != nil; {
			workFunc = workFunc()
		}
		w.Cancel()
		w.logger.Info("end")
		w.running = false
	}()
}

func (w *Work) Err() error {
	return w.tasks.Err()
}

func (w *Work) Running() bool {
	return w.running
}

func (w *Work) Tasks() Tasks {
	return w.tasks
}

func (w *Work) workStartJob() WorkFunc {
	var tasks []*config.Task
	for _, t := range w.taskConfigs {
		for _, s := range t.StartStates() {
			if s.Name == "JOB" && s.State == "START" {
				tasks = append(tasks, t)
				break
			}
		}
	}

	w.logger.Debug("tasks", "tasks", tasks)
	taskRunners := NewTaskRunners(tasks)
	taskRunners.Run() //Wait

	for _, tr := range taskRunners {
		w.tasks = append(w.tasks, tr)
	}

	return nil
}
