package config

type Job struct {
	Retry           *Retry       `json:retry,omitempty"`
	TaskDefault     *TaskDefault `json:"task_default,omitempty"`
	Tasks           []*Task      `json:"tasks"`
	FinishReportURL string       `json:"finish_report_url,omitempty"`
	//Tasks       map[string]*Task `json:"tasks"`
}
