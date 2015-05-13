package worker

import (
	"github.com/seanpont/assert"
	"golang.org/x/net/context"
	"gopkg.in/loom.v1/worker/config"
	"testing"
)

func TestJobSerialTasks(t *testing.T) {
	a := assert.Assert(t)
	ctx := context.Background()
	task1 := &config.Task{
		Name: "hello",
		Cmd:  "echo hello",
		When: "",
		Then: "hello",
	}
	task2 := &config.Task{
		Name: "world",
		Cmd:  "echo world",
		When: "hello",
		Then: "",
	}
	task3 := &config.Task{
		Name: "helloworld",
		Cmd:  "echo helloworld",
		When: "world",
	}
	jobConfig := &config.Job{}
	jobConfig.Tasks = append(jobConfig.Tasks, task1)
	jobConfig.Tasks = append(jobConfig.Tasks, task2)
	jobConfig.Tasks = append(jobConfig.Tasks, task3)

	job := NewJob(ctx, "id", jobConfig)
	job.Run()
	<-job.ctx.Done()

	a.Equal(len(job.Tasks), 3)

	for _, task := range job.Tasks {
		a.Equal(task.State(), TASK_STATE_DONE)
	}
}
