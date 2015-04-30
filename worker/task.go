package worker

import (
	"errors"
)

type Task interface {
	Name() string
	Ok() bool
	Err() error
	Output() string
}

type Tasks []Task

func (t Tasks) Err() error {
	return errors.New("TODO")
}
