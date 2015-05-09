package job

import (
	"github.com/looplab/fsm"
	"gopkg.in/loom.v1/log"
	"gopkg.in/loom.v1/worker/config"
	"os/exec"
)

type TaskRunner struct {
	job    *Job
	task   *config.Task
	err    error
	output string
	fsm    *fsm.FSM
	eventC chan string
	logger log.Logger
}

func NewTaskRunner(job *Job, task *config.Task) *TaskRunner {
	tr := &TaskRunner{
		job:    job,
		task:   task,
		eventC: make(chan string),
		logger: log.New("taskrunner#" + task.TaskName()),
	}
	tr_fsm := fsm.NewFSM(
		"init",
		fsm.Events{
			{Name: "run", Src: []string{"init"}, Dst: "process"},
			{Name: "success", Src: []string{"process"}, Dst: "success"},
			{Name: "cancel", Src: []string{"init", "process"}, Dst: "cancel"},
			{Name: "error", Src: []string{"process"}, Dst: "error"},
		},
		fsm.Callbacks{
			"process": func(e *fsm.Event) {
				err := tr.processing()
				if err != nil {
					tr.fsm.Event("error")
				}
				err = tr.fsm.Event("success")
				tr.logger.Debug("process... err:%v", err)
			},
			"success": func(e *fsm.Event) {
				tr.logger.Debug("success")
				tr.job.DoneTask(tr)
				tr.logger.Debug("success")
			},
			"error": func(e *fsm.Event) {
				tr.job.DoneTask(tr)
			},
			"cancel": func(e *fsm.Event) {
				tr.job.DoneTask(tr)
			},
		},
	)
	tr.fsm = tr_fsm
	go tr.eventListening()
	return tr
}

func (tr *TaskRunner) Run() {
	tr.eventC <- "run"
}

func (tr *TaskRunner) Cancel() {
	tr.eventC <- "cancel"
}

func (tr *TaskRunner) eventListening() {
	for {
		select {
		case event := <-tr.eventC:
			err := tr.fsm.Event(event)
			if err != nil {
				//TODO:
				tr.logger.Error("fsm event error: %v", err)
			}
		}
	}
}

func (tr *TaskRunner) processing() error {
	if tr.task.Cmd != "" {
		cmdstr := tr.task.Cmd
		cmd := exec.Command("bash", "-c", cmdstr)
		out, err := cmd.CombinedOutput()
		if err != nil {
			tr.err = err
		}
		tr.output = string(out)

		return err
	}
	return nil
}

func (tr *TaskRunner) TaskName() string {
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
