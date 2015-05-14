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

	tasks := make(Tasks)
	tasks["hello"] = NewMockTask("hello", TASK_STATE_DONE)
	tasks["world"] = NewMockTask("world", TASK_STATE_DONE)

	when := "hello==DONE"
	fTasks, err := f.Filter(when, tasks)
	if err != nil {
		t.Error(err)
	}

	a.Equal(len(fTasks), 1)
	a.Equal(fTasks["hello"], tasks["hello"])

	if _, ok := fTasks["hello"]; !ok {
		t.Error("this tasks is not hello")
	}
}

func TestTaskWhenFilterDefaultValue(t *testing.T) {
	a := assert.Assert(t)
	//f := &TaskWhenFilter{}
	f := taskWhenFilter

	tasks := make(Tasks)
	tasks["hello"] = NewMockTask("hello", TASK_STATE_DONE)
	tasks["world"] = NewMockTask("world", TASK_STATE_DONE)

	when := "hello"
	fTasks, err := f.Filter(when, tasks)
	if err != nil {
		t.Error(err)
	}

	a.Equal(len(fTasks), 1)
	a.Equal(fTasks["hello"], tasks["hello"])

	if _, ok := fTasks["hello"]; !ok {
		t.Error("this tasks is not hello")
	}
}
