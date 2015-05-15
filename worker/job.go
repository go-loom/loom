package worker

import (
	"fmt"
	"golang.org/x/net/context"
	"gopkg.in/loom.v1/log"
	"gopkg.in/loom.v1/worker/config"
)

type Job struct {
	ID                         string
	ctx                        context.Context
	cancelF                    context.CancelFunc
	config                     *config.Job
	Tasks                      Tasks
	changeTaskC                chan *TaskRunner
	doneTaskC                  chan *TaskRunner
	onTaskStateChangeHandelers []func(Task)
	logger                     log.Logger
}

func NewJob(ctx context.Context, id string, jobConfig *config.Job) *Job {
	_ctx, cf := context.WithCancel(ctx)
	job := &Job{
		ID:          id,
		ctx:         _ctx,
		cancelF:     cf,
		config:      jobConfig,
		Tasks:       make(map[string]Task),
		changeTaskC: make(chan *TaskRunner),
		doneTaskC:   make(chan *TaskRunner),
		logger:      log.New("job#" + id),
	}
	job.addTasks()

	go job.do()
	return job
}

func (job *Job) addTasks() {
	for _, task := range job.config.Tasks {
		job.Tasks[task.Name] = task
	}
}

func (job *Job) Run() {
	err := job.runTasks(&JobStartTask{})
	if err != nil {
		job.cancelF()
	}
}

func (job *Job) Done() <-chan struct{} {
	return job.ctx.Done()
}

func (job *Job) OnTaskDone(tr *TaskRunner) {
	job.doneTaskC <- tr
}

func (job *Job) OnTaskChanged(tr *TaskRunner) {
	job.changeTaskC <- tr
}

func (job *Job) OnTaskStateChange(handler func(Task)) {
	job.onTaskStateChangeHandelers = append(job.onTaskStateChangeHandelers, handler)
}

func (job *Job) do() {
L:
	for {
		select {
		case tr := <-job.doneTaskC:
			if job.logger.IsDebug() {
				job.logger.Debug("Recv done taskrunner name:%v state:%v", tr.TaskName(), tr.State())
			}
			job.Tasks[tr.TaskName()] = tr

			if job.isJobFinish() == true {
				break L
			}

			job.runTasks(tr)

		case tr := <-job.changeTaskC:
			for _, h := range job.onTaskStateChangeHandelers {
				h(tr)
			}
		}
	}

	job.cancelF()

	if job.logger.IsDebug() {
		job.logger.Debug("End do loop")
	}
}

func (job *Job) runTasks(task Task) error {
	matchTasks, notmatchTasks, err := taskRunFilter.Filter(task, job.config.Tasks)
	if err != nil {
		return err
	}
	for _, t := range matchTasks {
		tr := NewTaskRunner(job, t)
		tr.Run()

		if job.logger.IsDebug() {
			job.logger.Debug("Current Task name:%v state:%v -> Next task: %v", task.TaskName(), task.State(), tr.TaskName())
		}
	}

	for _, t := range notmatchTasks {
		tr := NewTaskRunner(job, t)
		tr.Cancel()

		if job.logger.IsDebug() {
			job.logger.Debug("Current Task name:%v state:%v -> Cancel task:%v", task.TaskName(), task.State(), tr.TaskName())
		}
	}

	if len(matchTasks) == 0 && len(notmatchTasks) == 0 {
		return fmt.Errorf("The task has no related next tasks (%v)", task.TaskName())
	}

	return nil
}

func (job *Job) isJobFinish() bool {
	numDoneT := 0
	for _, t := range job.Tasks {
		if t.State() == TASK_STATE_DONE || t.State() == TASK_STATE_ERROR || t.State() == TASK_STATE_CANCEL {
			numDoneT++
		}
	}
	if job.logger.IsDebug() {
		job.logger.Debug("isJobFinish numDoneT:%v tasks:%v", numDoneT, len(job.Tasks))
	}

	if numDoneT == len(job.Tasks) {
		return true
	}
	return false
}
