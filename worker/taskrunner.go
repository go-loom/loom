package worker

import (
	"fmt"
	"gopkg.in/loom.v1/worker/config"
	"os/exec"
	"sync"
)

type TaskRunner struct {
	task   *config.Task
	err    error
	output string
}

type TaskRunners struct {
	Job     Job
	Runners []*TaskRunner
}

func NewTaskRunners(tasks []*config.Task, job Job) *TaskRunners {
	trs := &TaskRunners{
		Job:     job,
		Runners: make([]*TaskRunner, 0, len(tasks)),
	}
	for _, t := range tasks {

		tr := &TaskRunner{
			task: t,
		}
		trs.Runners = append(trs.Runners, tr)
	}
	return trs
}

func (trs TaskRunners) Run() {
	var wg sync.WaitGroup
	wg.Add(len(trs.Runners))

	for _, tr := range trs.Runners {
		go func() {
			tr.Run(trs.Job)
			wg.Done()
		}()
	}
	wg.Wait()
}

func (tr *TaskRunner) Run(job Job) {
	if tr.task.Cmd != "" {
		cmdstr, err := tr.task.Read(tr.task.Cmd, job)
		if err != nil {
			tr.err = err
			return
		}
		cmd := exec.Command("bash", "-c", cmdstr)
		out, err := cmd.CombinedOutput()
		if err != nil {
			tr.err = err
		}
		fmt.Printf("out: %v\n", string(out))
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
	return tr.output
}
