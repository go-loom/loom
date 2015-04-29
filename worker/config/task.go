package config

type Task struct {
	Name      string
	Cmd       string
	When      string
	Then      string
	TaskState TaskState
	Timeout   string
	Retry     Retry
}

type TaskDefault struct {
	Retry   Retry
	Timeout string
	Vars    map[string]string
}

type TaskState struct {
	TaskName  string
	TaskState string
}
