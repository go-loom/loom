package config

import ()

type Task struct {
	Name    string
	Cmd     string
	When    interface{}
	Then    interface{}
	Timeout string
	Retry   Retry

	startState *TaskState
	endStates  *TaskState
}

type TaskDefault struct {
	Retry   Retry
	Timeout string
	Vars    map[string]string
}

type TaskState struct {
	Name  string
	State string
}

func (t *Task) StartState() *TaskState {
	if t.startState != nil {
		return t.startState
	}

	switch when := t.When.(type) {
	case string:
		t.startState = NewTaskState(when, "START")
		return t.startState
	case map[string]map[string]string:
		if s, ok := when["state"]; ok {
			for k, v := range s {
				t.startState = NewTaskState(k, v)
			}
		}
	}
	return t.startState
}

func NewTaskState(name, state string) *TaskState {
	ts := &TaskState{name, state}
	return ts
}
