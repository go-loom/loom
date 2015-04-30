package worker

import (
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
}

func NewWork(ctx context.Context, job Job, tasks []*config.Task) *Work {
	_ctx, cancelFunc := context.WithCancel(ctx)
	w := &Work{
		ctx:         _ctx,
		cancelFunc:  cancelFunc,
		Job:         job,
		taskConfigs: tasks,
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
		w.running = true
		for workFunc := w.workStartJob; workFunc != nil; {
			workFunc = workFunc()
		}
		w.Cancel()
		w.running = false
	}()
}

func (w *Work) Ok() bool {
	return false
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

	/*
		taskRunner := NewTaskRunner(tasks)
		taskRunner.Run() //Wait
		if taskRunner.Error() != nil {
		}
	*/

	return nil
}
