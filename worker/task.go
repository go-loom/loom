package worker

import (
	"time"
)

type Task interface {
	TaskName() string
	State() string
	Ok() bool
	Err() error
	Output() string
	StartEndTimes() []*time.Time
}
