package job

import (
	"golang.org/x/net/context"
	"gopkg.in/loom.v1/log"
	"gopkg.in/loom.v1/worker/config"
)

type Job struct {
	ctx       context.Context
	cancelF   context.CancelFunc
	config    *config.Worker
	tasks     map[string]Task
	doneTaskC chan *TaskRunner
	logger    log.Logger
}

//TODO: config.Worker -> config.Job
func NewJob(ctx context.Context, jobConfig *config.Worker) *Job {
	_ctx, cf := context.WithCancel(ctx)
	job := &Job{
		ctx:       _ctx,
		cancelF:   cf,
		config:    jobConfig,
		tasks:     make(map[string]Task),
		doneTaskC: make(chan *TaskRunner),
		logger:    log.New("job#" + jobConfig.Name),
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
	job.runNextTasks("JOB", "START")
}

func (job *Job) DoneTask(tr *TaskRunner) {
	job.doneTaskC <- tr
}

func (job *Job) do() {
	for {
		select {
		case tr := <-job.doneTaskC:
			if job.logger.IsDebug() {
				job.logger.Debug("Recv taskrunner done: %v", tr.TaskName())
			}
			job.tasks[tr.TaskName()] = tr
			job.logger.Debug("1")
			job.runNextTasks(tr.TaskName(), tr.State())
			job.logger.Debug("2")
			if job.finishJob() == true {
				job.logger.Debug("3")
				break
			}
			job.logger.Debug("4")
		}
	}
}

func (job *Job) runNextTasks(name, state string) error {
	job.logger.Debug("name:%v state:%v", name, state)
	var nextTasks []*config.Task
	nextTasks = job.config.FindTasksWhen(name, state, nextTasks)
	if job.logger.IsDebug() {
		job.logger.Debug("next tasks: %v", nextTasks)
	}
	for _, t := range nextTasks {
		tr := NewTaskRunner(job, t)
		tr.Run()

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
