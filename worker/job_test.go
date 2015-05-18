package worker

import (
	"github.com/seanpont/assert"
	"golang.org/x/net/context"
	c "gopkg.in/loom.v1/config"
	"testing"
)

func newTestJobRun(tasks []*c.Task, jobId string) *Job {
	ctx := context.Background()
	jobConfig := &c.Job{}
	jobConfig.Tasks = tasks

	job := NewJob(ctx, jobId, jobConfig)
	job.Run()
	<-job.ctx.Done()

	return job
}

func TestJobSerialTasks(t *testing.T) {
	a := assert.Assert(t)
	tasks := []*c.Task{
		&c.Task{
			Name: "hello",
			Cmd:  "echo hello",
			When: "",
		},
		&c.Task{
			Name: "world",
			Cmd:  "echo world",
			When: "hello",
		},
		&c.Task{
			Name: "helloworld",
			Cmd:  "echo helloworld",
			When: "world",
		},
	}

	job := newTestJobRun(tasks, "testjob1")
	a.Equal(len(job.Tasks), 3)

	for _, task := range job.Tasks {
		t.Logf("task name:%v", task.TaskName())
		a.Equal(task.State(), TASK_STATE_DONE)
	}
}

func TestJobHasErrorTask(t *testing.T) {
	a := assert.Assert(t)

	tasks := []*c.Task{
		&c.Task{
			Name: "hello",
			Cmd:  "echo hello",
			When: "",
		},
		&c.Task{
			Name: "world",
			Cmd:  "echo world ; exit 1",
			When: "hello",
		},
		&c.Task{
			Name: "helloworld",
			Cmd:  "echo helloworld",
			When: "world",
		},
	}

	job := newTestJobRun(tasks, "jobHasError")
	a.Equal(len(job.Tasks), 3)

	a.Equal(job.Tasks["hello"].State(), TASK_STATE_DONE)
	a.Equal(job.Tasks["world"].State(), TASK_STATE_ERROR)
	a.Equal(job.Tasks["helloworld"].State(), TASK_STATE_CANCEL)
}

func TestJobTaskHasErrTask(t *testing.T) {
	a := assert.Assert(t)

	tasks := []*c.Task{
		&c.Task{
			Name: "task1",
			Cmd:  "echo task1 ; exit 1",
			When: "JOB",
		},
		&c.Task{
			Name: "task2",
			Cmd:  "echo task2",
			When: "task1==DONE",
		},
		&c.Task{
			Name: "task3",
			Cmd:  "echo task3",
			When: "task1==ERROR",
		},
	}

	job := newTestJobRun(tasks, "jobErrTask")
	a.Equal(len(job.Tasks), 3)
	a.Equal(job.Tasks["task1"].State(), TASK_STATE_ERROR)
	a.Equal(job.Tasks["task2"].State(), TASK_STATE_CANCEL)
	a.Equal(job.Tasks["task3"].State(), TASK_STATE_DONE)
}
