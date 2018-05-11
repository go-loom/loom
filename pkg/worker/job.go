package worker

import (
	"context"
	"fmt"

	"github.com/go-loom/loom/pkg/config"
	"github.com/go-loom/loom/pkg/log"

	kitlog "github.com/go-kit/kit/log"
)

type Job struct {
	ID                         string
	ctx                        context.Context
	cancelF                    context.CancelFunc
	config                     *config.Job
	Tasks                      Tasks
	jobEndTasks                []*config.Task
	changeTaskC                chan *TaskRunner
	doneTaskC                  chan *TaskRunner
	onTaskStateChangeHandelers []func(Task)
	logger                     kitlog.Logger
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
		logger:      log.With(log.Logger, "job", id),
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
	task := NewJobTask("START")
	matchTasks, jobEndTasks, err := taskRunFilter.Filter(task, job.config.Tasks)
	if err != nil {
		log.Error(job.logger).Log("err", err)
		job.cancelF()
		return
	}

	job.jobEndTasks = jobEndTasks

	taskTemplateMap := job.Tasks.JSON()
	taskTemplateMap["JOB_ID"] = job.ID

	for _, t := range matchTasks {
		tr := NewTaskRunner(job, t, taskTemplateMap)
		tr.Run()
		log.Debug(job.logger).Log("task", task.TaskName(), "state", task.State(), "name", tr.TaskName)
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
			log.Debug(job.logger).Log("msg", "recv done", "name", tr.TaskName(), "state", tr.State())
			job.Tasks[tr.TaskName()] = tr

			if isFin, hasErr := job.isFinishTasks(); isFin == true {
				if hasErr {
					job.runTasks(NewJobTask(TASK_STATE_ERROR), job.jobEndTasks)
				} else {
					job.runTasks(NewJobTask(TASK_STATE_DONE), job.jobEndTasks)
				}
			} else {
				job.runTasks(tr)

			}

			if job.isFinishJob() == true {
				break L
			}

		case tr := <-job.changeTaskC:
			for _, h := range job.onTaskStateChangeHandelers {
				h(tr)
			}
		}
	}

	job.cancelF()

	log.Debug(job.logger).Log("msg", "End Doloop")
}

func (job *Job) runTasks(task Task, tasks ...[]*config.Task) error {
	var _tasks []*config.Task
	if len(tasks) > 0 {
		_tasks = tasks[0]
	} else {
		_tasks = job.config.Tasks
	}

	matchTasks, notmatchTasks, err := taskRunFilter.Filter(task, _tasks)
	if err != nil {
		return err
	}
	taskTemplateMap := job.Tasks.JSON()
	taskTemplateMap["JOB_ID"] = job.ID
	for _, t := range matchTasks {
		tr := NewTaskRunner(job, t, taskTemplateMap)
		tr.Run()
		log.Debug(job.logger).Log("task", task.TaskName(), "state", task.State(), "name", tr.TaskName())
	}

	for _, t := range notmatchTasks {
		tr := NewTaskRunner(job, t, taskTemplateMap)
		tr.Cancel()
		log.Debug(job.logger).Log("task", task.TaskName(), "state", task.State(), "name", tr.TaskName())
	}

	if len(matchTasks) == 0 && len(notmatchTasks) == 0 {
		return fmt.Errorf("The task has no related next tasks (%v)", task.TaskName())
	}

	return nil
}

func (job *Job) isFinishTasks() (bool, bool) {
	total := 0
	hasErr := false
	for _, t := range job.Tasks {
		for _, et := range job.jobEndTasks {
			if t.TaskName() == et.TaskName() {
				break
			}
		}

		if t.State() == TASK_STATE_DONE || t.State() == TASK_STATE_ERROR || t.State() == TASK_STATE_CANCEL {
			total++
		}
		if t.State() == TASK_STATE_ERROR {
			hasErr = true
		}
	}
	log.Debug(job.logger).Log("msg", "isFinishTasks", "total", total, "tasks", len(job.Tasks))

	if total == len(job.Tasks)-len(job.jobEndTasks) {
		return true, hasErr
	}

	return false, hasErr
}

func (job *Job) isFinishJob() bool {
	total := 0
	for _, t := range job.Tasks {
		if t.State() == TASK_STATE_DONE || t.State() == TASK_STATE_ERROR || t.State() == TASK_STATE_CANCEL {
			total++
		}
	}
	log.Debug(job.logger).Log("msg", "isFinishJob", "total", total, "tasks", len(job.Tasks))

	if total == len(job.Tasks) {
		return true
	}
	return false
}
