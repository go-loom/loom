package worker

import (
	c "github.com/go-loom/loom/config"
	"strings"
)

type TaskRunFilter struct{}

var (
	taskRunFilter *TaskRunFilter
)

func init() {
	taskRunFilter = &TaskRunFilter{}
}

func (f *TaskRunFilter) Name() string {
	return "when"
}

func (f *TaskRunFilter) Filter(task Task, tasks []*c.Task) (matched []*c.Task, notmatched []*c.Task, err error) {

	for _, t := range tasks {
		exprs, err := parseExprs([]string{t.When})
		if err != nil {
			return nil, nil, err
		}
		for _, expr := range exprs {
			switch expr.operator {
			case EQ: //==
				if task.TaskName() == expr.key {
					if task.State() == strings.ToUpper(expr.value) {
						matched = append(matched, t)
					} else {
						notmatched = append(notmatched, t)
					}
				}
			case NOTEQ: //!=
				if task.TaskName() == expr.key {
					if task.State() != strings.ToUpper(expr.value) {
						matched = append(matched, t)
					} else {
						notmatched = append(notmatched, t)
					}
				}
			}
		}
	}

	return
}
