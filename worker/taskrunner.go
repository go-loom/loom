package worker

import (
	"gopkg.in/loom.v1/worker/config"
)

type TaskRunner struct {
	task *config.Task
}

type TaskRunners []*TaskRunner

func NewTaskRunners(tasks []*config.Task) TaskRunners {
	trs := make(TaskRunners, 0, len(tasks))
	for _, t := range tasks {

		tr := &TaskRunner{
			task: t,
		}
		trs = append(trs, tr)
	}
	return trs
}

func (trs TaskRunners) Run() {
	go func() {

	}()
}

func (tr *TaskRunner) Name() string {
	return tr.task.Name
}

func (tr *TaskRunner) Ok() bool {
	return false
}

func (tr *TaskRunner) Err() error {
	return nil
}

func (tr *TaskRunner) Output() string {
	return ""
}
