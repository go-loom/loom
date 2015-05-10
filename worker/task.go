package worker

type Task interface {
	TaskName() string
	State() string
	Ok() bool
	Err() error
	Output() string
}
