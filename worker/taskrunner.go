package worker

import (
	"gopkg.in/loom.v1/worker/config"
	"os/exec"
)

type TaskRunner struct {
	task   *config.Task
	err    error
	output string
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

func (tr *TaskRunner) Run() {
	if tr.task.Cmd != "" {
		cmd := exec.Command("bash", "-c", tr.task.Cmd)
		out, err := cmd.CombinedOutput()
		if err != nil {
			tr.err = err
		}
		tr.output = string(out)
	}
}

func (tr *TaskRunner) Name() string {
	return tr.task.Name
}

func (tr *TaskRunner) Ok() bool {
	if tr.err == nil {
		return true
	}
	return false
}

func (tr *TaskRunner) Err() error {
	return tr.err
}

func (tr *TaskRunner) Output() string {
	return ""
}
