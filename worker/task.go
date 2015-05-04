package worker

type Task interface {
	Name() string
	Ok() bool
	Err() error
	Output() string
}

type Tasks []Task

func (tasks Tasks) Err() error {
	for _, t := range tasks {
		if t.Err() != nil {
			return t.Err()
		}
	}
	return nil
}

func (tasks Tasks) MapInfo() []map[string]interface{} {
	var result []map[string]interface{}
	for _, t := range tasks {
		taskMap := map[string]interface{}{
			"name":   t.Name(),
			"ok":     t.Ok(),
			"err":    "",
			"output": t.Output(),
		}
		if t.Err() != nil {
			taskMap["err"] = t.Err().Error()
		}
		result = append(result, taskMap)
	}
	return result
}
