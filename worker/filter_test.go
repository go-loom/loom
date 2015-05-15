package worker

import (
	"github.com/seanpont/assert"
	"gopkg.in/loom.v1/worker/config"
	"testing"
)

type MockTask struct {
	config.Task
	name  string
	state string
}

func (t *MockTask) TaskName() string {
	return t.name
}

func (t *MockTask) State() string {
	return t.state
}

func NewMockTask(name, state string) Task {
	return &MockTask{name: name, state: state}
}

func TestTaskRunFilter(t *testing.T) {
	a := assert.Assert(t)
	f := &TaskRunFilter{}

	var taskConfigs []*config.Task
	taskConfigs = append(taskConfigs, &config.Task{Name: "world", When: "hello==DONE"})

	matched, notmatched, err := f.Filter(NewMockTask("hello", TASK_STATE_DONE), taskConfigs)
	if err != nil {
		t.Error(err)
	}

	a.Equal(len(matched), 1)
	a.Equal(len(notmatched), 0)

	found := false
	for _, t := range matched {
		if t.TaskName() == "world" {
			found = true
		}
	}

	if !found {
		t.Error("this tasks has not hello task")
	}
}

func TestTaskRunFilterDefaultValue(t *testing.T) {
	a := assert.Assert(t)
	f := taskRunFilter

	var taskConfigs []*config.Task
	taskConfigs = append(taskConfigs, &config.Task{Name: "world", When: "hello"})

	matched, notmatched, err := f.Filter(NewMockTask("hello", TASK_STATE_DONE), taskConfigs)
	if err != nil {
		t.Error(err)
	}

	a.Equal(len(matched), 1)
	a.Equal(len(notmatched), 0)

	found := false
	for _, t := range matched {
		if t.TaskName() == "world" {
			found = true
		}
	}

	if !found {
		t.Error("this tasks has not hello task")
	}

}

func TestTaskRunFilterJOBTask(t *testing.T) {
	a := assert.Assert(t)
	f := taskRunFilter

	var taskConfigs []*config.Task
	taskConfigs = append(taskConfigs, &config.Task{Name: "world", When: "JOB"})

	matched, notmatched, err := f.Filter(NewMockTask("JOB", "START"), taskConfigs)
	if err != nil {
		t.Error(err)
	}

	a.Equal(len(matched), 1)
	a.Equal(len(notmatched), 0)

	found := false
	for _, t := range matched {
		if t.TaskName() == "world" {
			found = true
		}
	}

	if !found {
		t.Error("this tasks has not hello task")
	}

}

func TestTaskRunFilterJOBTask2(t *testing.T) {
	a := assert.Assert(t)
	f := taskRunFilter

	var taskConfigs []*config.Task
	taskConfigs = append(taskConfigs, &config.Task{Name: "world", When: ""})

	matched, notmatched, err := f.Filter(NewMockTask("JOB", "START"), taskConfigs)
	if err != nil {
		t.Error(err)
	}

	a.Equal(len(matched), 1)
	a.Equal(len(notmatched), 0)

	found := false
	for _, t := range matched {
		if t.TaskName() == "world" {
			found = true
		}
	}

	if !found {
		t.Error("this tasks has not hello task")
	}

}
