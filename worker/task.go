package worker

import (
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

func (tasks Tasks) JSON() (result []map[string]interface{}) {
	for _, t := range tasks {

		taskMap := map[string]interface{}{
			"name":   t.TaskName(),
			"ok":     t.Ok(),
			"err":    "",
			"output": t.Output(),
			"state":  t.State(),
		}

		ts := t.StartEndTimes()
		if len(ts) >= 1 {
			st := ts[0]
			taskMap["started"] = st.UTC().Format(time.RFC3339)
		}
		if len(ts) >= 2 {
			et := ts[1]
			taskMap["ended"] = et.UTC().Format(time.RFC3339)
		}

		if t.Err() != nil {
			taskMap["err"] = t.Err().Error()
		}

		result = append(result, taskMap)
	}

	return
}
