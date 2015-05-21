package config

type Job struct {
	Retry       *Retry       `json:retry,omitempty"`
	TaskDefault *TaskDefault `json:"task_default,omitempty"`
	Tasks       []*Task      `json:"tasks"`
	//Tasks       map[string]*Task `json:"tasks"`
}
