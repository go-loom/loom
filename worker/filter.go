package worker

import (
	"strings"
)

type TaskWhenFilter struct{}

var (
	taskWhenFilter *TaskWhenFilter
)

func init() {
	taskWhenFilter = &TaskWhenFilter{}
}

func (f *TaskWhenFilter) Name() string {
	return "when"
}

func (f *TaskWhenFilter) Filter(when string, tasks Tasks) (Tasks, error) {
	exprs, err := parseExprs([]string{when})
	if err != nil {
		return nil, err
	}

	filteredTasks := make(Tasks)

	for _, expr := range exprs {
		if t, ok := tasks[expr.key]; ok {
			value := strings.ToUpper(expr.value)
			if value == "" && t.State() == TASK_STATE_DONE {
				filteredTasks[t.TaskName()] = t
			} else if value == t.State() {
				filteredTasks[t.TaskName()] = t
			}
		}
	}

	return filteredTasks, nil
}
