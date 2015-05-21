package server

type TaskResults struct {
	WorkerId string
	Tasks    map[string]interface{}
}

func NewTaskResults(workerId string, tasks map[string]interface{}) *TaskResults {
	tr := &TaskResults{
		WorkerId: workerId,
		Tasks:    tasks,
	}
	return tr
}
