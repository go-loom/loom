package worker

type Job map[string]interface{}

func NewJob(id string, values map[string]interface{}) Job {
	job := Job{}
	for k, v := range values {
		job[k] = v
	}
	job["ID"] = id
	return job
}
