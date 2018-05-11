package worker

import (
	"github.com/go-loom/loom/pkg/config"
)

type JobMessage struct {
	Tasks []*config.Task `json:tasks`
}
