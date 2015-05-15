package worker

import (
	"gopkg.in/loom.v1/worker/config"
)

type TaskRunFilter struct{}
type TaskCancelFilter struct{}

var (
	taskRunFilter    *TaskRunFilter
	taskCancelFilter *TaskCancelFilter
)

func init() {
	taskRunFilter = &TaskRunFilter{}
	taskCancelFilter = &TaskCancelFilter{}
}

func (f *TaskRunFilter) Name() string {
	return "when"
}

func (f *TaskRunFilter) Filter(task Task, tasks []*config.Task) (map[string]*config.Task, error) {
	fTasks := make(map[string]*config.Task)

	for _, t := range tasks {
		exprs, err := parseExprs([]string{t.When})
		if err != nil {
			return nil, err
		}
		for _, expr := range exprs {
			switch expr.operator {
			case EQ: //==
				if task.TaskName() == expr.key && task.State() == expr.value {
					fTasks[t.Name] = t
				}
			case NOTEQ: //!=
				if task.TaskName() == expr.key && task.State() != expr.value {
					fTasks[t.Name] = t
				}
			}
		}
	}

	return fTasks, nil
}

func (f *TaskCancelFilter) Name() string {
	return "when"
}

func (f *TaskCancelFilter) Filter(task Task, tasks []*config.Task) (map[string]*config.Task, error) {
	fTasks := make(map[string]*config.Task)

	for _, t := range tasks {
		exprs, err := parseExprs([]string{t.When})
		if err != nil {
			return nil, err
		}
		for _, expr := range exprs {
			switch expr.operator {
			case EQ: //==
				if task.TaskName() == expr.key {
					if task.State() == TASK_STATE_ERROR || task.State() == TASK_STATE_CANCEL {
						fTasks[t.Name] = t
					}
				}
			}
		}
	}

	return fTasks, nil
}
