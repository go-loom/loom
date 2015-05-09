package job

import (
	"github.com/seanpont/assert"
	"golang.org/x/net/context"
	"gopkg.in/loom.v1/worker/config"
	"testing"
	"time"
)

func TestNewTaskRunner(t *testing.T) {
	a := assert.Assert(t)
	ctx := context.Background()
	task := &config.Task{
		Name: "helloworld",
		Cmd:  "echo helloworld",
	}
	jobConfig := &config.Worker{
		Name: "test",
	}
	jobConfig.Tasks = append(jobConfig.Tasks, task)

	job := NewJob(ctx, jobConfig)

	tr := NewTaskRunner(job, task)
	tr.Run()

	time.Sleep(1 * time.Second)
	a.Equal(tr.fsm.Current(), "success")

}
