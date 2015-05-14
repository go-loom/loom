package config

import (
	"time"
)

type Task struct {
	templateReader
	Name    string
	Cmd     string
	HTTP    *HTTP
	When    string
	Timeout string
	Retry   Retry
}

type TaskDefault struct {
	templateReader
	Retry   Retry
	Timeout string
	Vars    map[string]string
}

func (t *Task) TaskName() string {
	return t.Name
}

func (t *Task) Ok() bool {
	return false
}

func (t *Task) Err() error {
	return nil
}

func (t *Task) Output() string {
	return ""
}

func (t *Task) State() string {
	return "INIT"
}

func (t *Task) StartEndTimes() []*time.Time {
	return []*time.Time{}
}
