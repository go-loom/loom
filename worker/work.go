package worker

import (
	"github.com/mgutz/logxi/v1"
	"golang.org/x/net/context"
	"gopkg.in/loom.v1/worker/config"
)

type WorkFunc func() WorkFunc

type Work struct {
	ctx          context.Context
	cancelFunc   context.CancelFunc
	Job          Job
	workerConfig *config.Worker
	tasks        Tasks
	taskRunners  *TaskRunners
	running      bool
	logger       log.Logger
}

func NewWork(ctx context.Context, job Job, workerConfig *config.Worker) *Work {
	_ctx, cancelFunc := context.WithCancel(ctx)
	w := &Work{
		ctx:          _ctx,
		cancelFunc:   cancelFunc,
		Job:          job,
		workerConfig: workerConfig,
		logger:       log.New("work#" + job["ID"].(string)),
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
		for workFunc := w.workJobStart; workFunc != nil; {
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

func (w *Work) workJobStart() WorkFunc {

	var tasks []*config.Task
	for _, t := range w.workerConfig.Tasks {
		for _, s := range t.StartStates() {
			if s.Name == "JOB" && s.State == "START" {
				tasks = append(tasks, t)
				break
			}
		}
	}

	if len(tasks) <= 0 {
		return w.workJobDone
	}

	if w.logger.IsDebug() {
		for _, t := range tasks {
			w.logger.Debug("workJobStart", "task", t.Name, "when", t.StartStates())
		}
	}

	w.taskRunners = NewTaskRunners(tasks, w.Job)
	w.taskRunners.Run() //Wait

	for _, tr := range w.taskRunners.Runners {
		w.tasks = append(w.tasks, tr)
	}

	return w.workAfterTask
}

func (w *Work) workAfterTask() WorkFunc {
	var tasks []*config.Task

	for _, tr := range w.taskRunners.Runners {
		for _, state := range tr.task.EndStates() {
			if w.logger.IsDebug() {
				w.logger.Debug("workAfterTask", "name", state.Name, "state", state.State)
			}
			tasks = w.workerConfig.FindTasksWhen(state.Name, state.State, tasks)
		}
	}

	if w.logger.IsDebug() {
		for _, t := range tasks {
			w.logger.Debug("workAfterTask", "task", t.Name, "when", t.StartStates())
		}
	}

	if len(tasks) <= 0 {
		return w.workJobDone
	}

	w.taskRunners = NewTaskRunners(tasks, w.Job)
	w.taskRunners.Run() //Wait

	for _, tr := range w.taskRunners.Runners {
		w.tasks = append(w.tasks, tr)
	}

	return w.workAfterTask
}

func (w *Work) workJobDone() WorkFunc {
	w.logger.Debug("workJobDone")

	var tasks []*config.Task
	tasks = w.workerConfig.FindTasksWhen("JOB", "DONE", tasks)
	if len(tasks) <= 0 {
		return nil
	}

	w.taskRunners = NewTaskRunners(tasks, w.Job)
	w.taskRunners.Run() //Wait

	for _, tr := range w.taskRunners.Runners {
		w.tasks = append(w.tasks, tr)
	}

	return nil
}
