package job

import (
	"golang.org/x/net/context"
	"gopkg.in/loom.v1/worker/config"
)

type Job struct {
	ctx         context.Context
	config      *config.Worker
	tasks       map[string]Task
	taskEdges   map[string]map[string]struct{}
	taskRunners map[string]*TaskRunner
	doneTaskC   chan *TaskRunner
}

//TODO: config.Worker -> config.Job
func NewJob(ctx context.Context, jobConfig *config.Worker) *Job {
	job := &Job{
		ctx:         ctx,
		config:      jobConfig,
		tasks:       make(map[string]Task),
		taskRunners: make(map[string]*TaskRunner),
		taskEdges:   make(map[string]map[string]struct{}),
		doneTaskC:   make(chan *TaskRunner),
	}
	job.addTasks()

	go job.do()
	return job
}

func (job *Job) addTasks() {
	for _, task := range job.config.Tasks {
		job.tasks[task.Name] = task

		//Add NewTaskRunner
		job.taskRunners[task.Name] = NewTaskRunner(job, task)
	}
}

func (job *Job) DoneTask(tr *TaskRunner) {
	job.doneTaskC <- tr
}

func (job *Job) do() {
	for {
		select {
		case tr := <-job.doneTaskC:
			job.tasks[tr.TaskName()] = tr
		}
	}
}
