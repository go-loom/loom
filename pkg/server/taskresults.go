package server

type TaskResults struct {
	WorkerId string                 `json:"worker"`
	Tasks    map[string]interface{} `json:"tasks"`
}

func NewTaskResults(workerId string, tasks map[string]interface{}) *TaskResults {
	tr := &TaskResults{
		WorkerId: workerId,
		Tasks:    tasks,
	}
	return tr
}
