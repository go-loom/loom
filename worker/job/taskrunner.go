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
	stateC chan string
	logger log.Logger
}

func NewTaskRunner(job *Job, task *config.Task) *TaskRunner {
	tr := &TaskRunner{
		job:    job,
		task:   task,
		eventC: make(chan string),
		stateC: make(chan string),
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
			"enter_state": func(e *fsm.Event) {
				tr.logger.Debug("enter state:%v", e.Dst)
				tr.stateC <- e.Dst
				tr.logger.Debug("enter state:%v", e.Dst)
			},
		},
	)
	tr.fsm = tr_fsm
	go tr.stateListening()
	go tr.eventListening()

	tr.logger.Info("New task runner")
	return tr
}

func (tr *TaskRunner) Run() {
	tr.eventC <- "run"
}

func (tr *TaskRunner) Cancel() {
	tr.eventC <- "cancel"
}

func (tr *TaskRunner) stateListening() {
	tr.logger.Info("start stateListening")
	for {
		select {
		case state := <-tr.stateC:
			tr.logger.Debug("stateC:%v", state)
			if state == "process" {
				err := tr.processing()
				if err != nil {
					tr.eventC <- "error"
				} else {
					tr.eventC <- "success"
				}
			} else {
				tr.job.DoneTask(tr)
			}
		}
	}
	tr.logger.Info("end stateListening")
}

func (tr *TaskRunner) eventListening() {
	for {
		select {
		case event := <-tr.eventC:
			err := tr.fsm.Event(event)
			tr.logger.Debug("eventListening event:%v", event)
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

func (tr *TaskRunner) State() string {
	return tr.fsm.Current()
}

func (tr *TaskRunner) Err() error {
	return tr.err
}

func (tr *TaskRunner) Output() string {
	return tr.output
}
