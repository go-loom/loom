package worker

import (
	"golang.org/x/net/context"
	"gopkg.in/loom.v1/log"
	"gopkg.in/loom.v1/worker/config"
)

type Job struct {
	ID        string
	ctx       context.Context
	cancelF   context.CancelFunc
	config    *config.Job
	tasks     map[string]Task
	doneTaskC chan *TaskRunner
	logger    log.Logger
}

func NewJob(ctx context.Context, id string, jobConfig *config.Job) *Job {
	_ctx, cf := context.WithCancel(ctx)
	job := &Job{
		ID:        id,
		ctx:       _ctx,
		cancelF:   cf,
		config:    jobConfig,
		tasks:     make(map[string]Task),
		doneTaskC: make(chan *TaskRunner),
		logger:    log.New("job#" + id),
	}
	job.addTasks()

	go job.do()
	return job
}

func (job *Job) addTasks() {
	for _, task := range job.config.Tasks {
		job.tasks[task.Name] = task
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

func (job *Job) DoneTask(tr *TaskRunner) {
	job.doneTaskC <- tr
}

func (job *Job) TasksMapInfo() []map[string]interface{} {
	var result []map[string]interface{}
	for _, t := range job.tasks {
		job.logger.Debug("tasks map name:%v  output:%v", t.TaskName(), t.Output())
		taskMap := map[string]interface{}{
			"name":   t.TaskName(),
			"ok":     t.Ok(),
			"err":    "",
			"output": t.Output(),
			"state":  t.State(),
		}
		if t.Err() != nil {
			taskMap["err"] = t.Err().Error()
		}
		result = append(result, taskMap)
	}
	return result
}

func (job *Job) do() {
	for {
		select {
		case tr := <-job.doneTaskC:
			if job.logger.IsDebug() {
				job.logger.Debug("Done taskrunner: %v ", tr.TaskName())
			}
			job.tasks[tr.TaskName()] = tr
			job.walkNextTasks(tr.TaskName(), tr.State(), func(ntr *TaskRunner) {
				if tr.State() == "ERROR" || tr.State() == "CANCEL" {
					ntr.Cancel()
				} else {
					ntr.Run()
				}
			})
			if job.finishJob() == true {
				break
			}
		}
	}
}

func (job *Job) walkNextTasks(name, state string, taskRunnerFunc func(*TaskRunner)) error {

	var nextTasks []*config.Task
	nextTasks = job.config.FindTasksWhen(name, state, nextTasks)

	for _, t := range nextTasks {
		tr := NewTaskRunner(job, t)

		if job.logger.IsDebug() {
			job.logger.Debug("name: %v state: %v next task: %v", name, state, tr.TaskName())
		}

		taskRunnerFunc(tr)

		if job.logger.IsDebug() {
			job.logger.Debug("Run taskrunner name:%v", tr.TaskName())
		}
	}
	return nil
}

func (job *Job) finishJob() bool {
	numDoneT := 0
	for _, t := range job.tasks {
		if t.State() == "DONE" || t.State() == "ERROR" || t.State() == "CANCEL" {
			numDoneT++
		}
	}
	if numDoneT == len(job.tasks) {
		job.cancelF()
		return true
	}
	return false

}
