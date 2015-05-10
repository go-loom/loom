package config

type Job struct {
	Tasks []*Task
}

func (j *Job) FindTasksWhen(name, state string, tasks []*Task) []*Task {

	for _, t := range j.Tasks {
		for _, s := range t.StartStates() {
			if s.Name == name && s.State == state {
				tasks = append(tasks, t)
				break
			}
		}
	}
	return tasks
}
