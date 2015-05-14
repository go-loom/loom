package worker

import (
	"gopkg.in/loom.v1/worker/config"
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

func (f *TaskWhenFilter) Filter(task Task, tasks []*config.Task) (map[string]*config.Task, error) {
	fTasks := make(map[string]*config.Task)

	for _, t := range tasks {
		exprs, err := parseExprs([]string{t.When})
		if err != nil {
			return nil, err
		}
		for _, expr := range exprs {
			switch expr.operator {
			case 0: //==
				if task.TaskName() == expr.key && task.State() == expr.value {
					fTasks[t.Name] = t
				}
			case 1: //!=
				if task.TaskName() == expr.key && task.State() != expr.value {
					fTasks[t.Name] = t
				}
			}
		}
	}

	return fTasks, nil
}
