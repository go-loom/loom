package config

import (
	"time"
)

type Task struct {
	templateReader
	Name    string `json:"name"`
	Cmd     string `json:"cmd,omitempty"`
	HTTP    *HTTP  `json:"http,omitempty"`
	When    string `json:"when,omitempty"`
	Timeout string `json:"timeout,omitempty"`
	Retry   Retry  `json:"retry,omitempty"`
}

type TaskDefault struct {
	templateReader
	Retry   Retry             `json:"retry,omitempty"`
	Timeout string            `json:"timeout,omitempty"`
	Vars    map[string]string `json:"vars,omitempty"`
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
