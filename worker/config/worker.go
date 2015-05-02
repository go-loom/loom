package config

type Worker struct {
	Name    string
	Version string
	Tasks   []*Task
}

func (w *Worker) FindTasksWhen(name, state string, tasks []*Task) []*Task {

	for _, t := range w.Tasks {
		for _, s := range t.StartStates() {
			if s.Name == name && s.State == state {
				tasks = append(tasks, t)
				break
			}
		}
	}
	return tasks
}
