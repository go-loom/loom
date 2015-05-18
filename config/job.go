package config

type Job struct {
	TaskDefault *TaskDefault `json:"task_default,omitempty"`
	Tasks       []*Task      `json:"tasks,omitempty"`
}
