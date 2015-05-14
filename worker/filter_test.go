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

func TestTaskWhenFilter(t *testing.T) {
	a := assert.Assert(t)
	f := &TaskWhenFilter{}

	var taskConfigs []*config.Task
	taskConfigs = append(taskConfigs, &config.Task{Name: "world", When: "hello==DONE"})

	fTasks, err := f.Filter(NewMockTask("hello", TASK_STATE_DONE), taskConfigs)
	if err != nil {
		t.Error(err)
	}

	a.Equal(len(fTasks), 1)

	if _, ok := fTasks["world"]; !ok {
		t.Error("this tasks is not hello")
	}
}

func TestTaskWhenFilterDefaultValue(t *testing.T) {
	a := assert.Assert(t)
	//f := &TaskWhenFilter{}
	f := taskWhenFilter

	var taskConfigs []*config.Task
	taskConfigs = append(taskConfigs, &config.Task{Name: "world", When: "hello"})

	fTasks, err := f.Filter(NewMockTask("hello", TASK_STATE_DONE), taskConfigs)
	if err != nil {
		t.Error(err)
	}

	a.Equal(len(fTasks), 1)

	if _, ok := fTasks["world"]; !ok {
		t.Error("this tasks is not hello")
	}
}
