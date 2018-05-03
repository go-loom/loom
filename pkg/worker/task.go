package worker

import (
	"github.com/go-loom/loom/pkg/config"

	"time"
)

type Task interface {
	TaskName() string
	State() string
	Ok() bool
	Err() error
	Output() string
	StartEndTimes() []*time.Time
}

type Tasks map[string]Task

func (tasks Tasks) JSON() (result map[string]interface{}) {
	result = make(map[string]interface{})
	for _, t := range tasks {

		taskMap := map[string]interface{}{
			"name":   t.TaskName(),
			"err":    "",
			"ok":     t.Ok(),
			"output": t.Output(),
			"state":  t.State(),
		}

		ts := t.StartEndTimes()
		if len(ts) >= 1 {
			st := ts[0]
			taskMap["started"] = st
		}
		if len(ts) >= 2 {
			et := ts[1]
			taskMap["ended"] = et
		}

		if t.Err() != nil {
			taskMap["err"] = t.Err().Error()
		}

		result[t.TaskName()] = taskMap
	}

	return
}

type JobTask struct {
	config.Task
	state string
}

func NewJobTask(state string) *JobTask {
	t := &JobTask{
		state: state,
	}
	return t
}

func (t *JobTask) TaskName() string {
	return "JOB"
}
func (t *JobTask) State() string {
	return t.state
}
