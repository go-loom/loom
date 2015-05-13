package worker

import (
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
	job.walkNextTasks("JOB", "START", func(tr *TaskRunner) {
		tr.Run()
	})
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
	for {
		select {
		case tr := <-job.doneTaskC:
			if job.logger.IsDebug() {
				job.logger.Debug("Done taskrunner: %v ", tr.TaskName())
			}
			job.Tasks[tr.TaskName()] = tr
			job.walkNextTasks(tr.TaskName(), tr.State(), func(ntr *TaskRunner) {
				if tr.State() == TASK_STATE_ERROR || tr.State() == TASK_STATE_CANCEL {
					ntr.Cancel()
				} else {
					ntr.Run()
				}
			})
			if job.finishJob() == true {
				break
			}
		case tr := <-job.changeTaskC:
			for _, h := range job.onTaskStateChangeHandelers {
				h(tr)
			}
		}
	}
}

func (job *Job) walkNextTasks(name, state string, taskRunnerFunc func(*TaskRunner)) error {

	var nextTasks []*config.Task
	nextTasks = job.config.FindTasksWhen(name, state, nextTasks)

	for _, t := range nextTasks {
		tr := NewTaskRunner(job, t)
		taskRunnerFunc(tr)

		if job.logger.IsDebug() {
			job.logger.Debug("Current Taskname: %v state: %v Next task: %v", name, state, tr.TaskName())
		}

	}
	return nil
}

func (job *Job) finishJob() bool {
	numDoneT := 0
	for _, t := range job.Tasks {
		if t.State() == TASK_STATE_DONE || t.State() == TASK_STATE_ERROR || t.State() == TASK_STATE_CANCEL {
			numDoneT++
		}
	}
	if numDoneT == len(job.Tasks) {
		job.cancelF()
		return true
	}
	return false
}
