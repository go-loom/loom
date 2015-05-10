package job

import (
	"github.com/seanpont/assert"
	"golang.org/x/net/context"
	"gopkg.in/loom.v1/worker/config"
	"testing"
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
	tr := job.taskRunners["helloworld"]
	tr.Run()

	<-job.ctx.Done()

	a.Equal(tr.fsm.Current(), "success")

}
