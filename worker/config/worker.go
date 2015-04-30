package config

type Worker struct {
	Name    string
	Version string
	Tasks   []*Task
}
