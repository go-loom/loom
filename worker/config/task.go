package config

import (
	"fmt"
)

type Task struct {
	Name    string
	Cmd     string
	When    interface{}
	Then    interface{}
	Timeout string
	Retry   Retry

	startStates []*TaskState
	endStates   []*TaskState
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

func (t *Task) StartStates() []*TaskState {
	if len(t.startStates) > 0 {
		return t.startStates
	}

	switch when := t.When.(type) {
	case string:
		var ts *TaskState
		if when == "" {
			ts = NewTaskState("JOB", "START")
		} else {
			ts = NewTaskState(when, "DONE")
		}
		t.startStates = append(t.startStates, ts)
	case map[interface{}]interface{}:
		if s, ok := when["state"]; ok {
			for k, v := range s.(map[interface{}]interface{}) {
				ts := NewTaskState(k.(string), v.(string))
				t.startStates = append(t.startStates, ts)
			}
		}
	default:
		fmt.Println("ERROR") //TODO: raise panic!
	}
	return t.startStates
}

func (t *Task) EndStates() []*TaskState {
	if len(t.endStates) > 0 {
		return t.endStates
	}

	switch then := t.Then.(type) {
	case string:
		var ts *TaskState
		if then == "" {
			ts = NewTaskState("JOB", "DONE")
		} else {
			ts = NewTaskState(then, "DONE")
		}
		t.endStates = append(t.endStates, ts)
	case map[interface{}]interface{}:
		if s, ok := then["state"]; ok {
			for k, v := range s.(map[interface{}]interface{}) {
				ts := NewTaskState(k.(string), v.(string))
				t.endStates = append(t.endStates, ts)
			}
		}

	default:
		fmt.Println("ERROR")
	}
	return t.endStates
}

func NewTaskState(name, state string) *TaskState {
	ts := &TaskState{name, state}
	return ts
}
