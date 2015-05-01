package worker

import ()

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
