package worker

import (
	"fmt"
	"github.com/seanpont/assert"
	"golang.org/x/net/context"
	"gopkg.in/loom.v1/worker/config"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewTaskRunner(t *testing.T) {
	a := assert.Assert(t)
	ctx := context.Background()
	task := &config.Task{
		Name: "helloworld",
		Cmd:  "echo helloworld",
	}
	jobConfig := &config.Job{}
	jobConfig.Tasks = append(jobConfig.Tasks, task)

	job := NewJob(ctx, "id", jobConfig)
	tr := NewTaskRunner(job, task)
	tr.Run()

	<-job.ctx.Done()

	a.Equal(tr.fsm.Current(), "DONE")
	a.Equal(tr.State(), "DONE")
	t.Logf("task.output:%v", tr.Output())

}

func TestTaskRunnerHTTPGet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	a := assert.Assert(t)
	ctx := context.Background()
	task := &config.Task{
		Name: "helloworld",
		HTTP: &config.HTTP{
			URL:    ts.URL,
			Method: "GET",
			Data: map[string]string{
				"p1": "1",
				"p2": "2",
			},
		},
	}
	jobConfig := &config.Job{}
	jobConfig.Tasks = append(jobConfig.Tasks, task)

	job := NewJob(ctx, "id", jobConfig)
	tr := NewTaskRunner(job, task)

	tr.Run()

	<-job.ctx.Done()

	a.Equal(tr.fsm.Current(), "DONE")
	a.Equal(tr.State(), "DONE")
	t.Logf("task.output:%v", tr.Output())

}

func TestTaskRunnerHTTPPost(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	a := assert.Assert(t)
	ctx := context.Background()
	task := &config.Task{
		Name: "helloworld",
		HTTP: &config.HTTP{
			URL:    ts.URL,
			Method: "POST",
			Data: map[string]string{
				"p1": "1",
				"p2": "2",
			},
		},
	}
	jobConfig := &config.Job{}
	jobConfig.Tasks = append(jobConfig.Tasks, task)

	job := NewJob(ctx, "id", jobConfig)
	tr := NewTaskRunner(job, task)

	tr.Run()

	<-job.ctx.Done()

	a.Equal(tr.fsm.Current(), "DONE")
	a.Equal(tr.State(), "DONE")
	t.Logf("task.output:%v", tr.Output())

}
