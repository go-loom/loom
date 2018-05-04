package worker

import (
	"github.com/go-loom/loom/pkg/config"

	"github.com/seanpont/assert"

	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"path/filepath"
	"testing"
)

func TestNewTaskRunner(t *testing.T) {
	a := assert.Assert(t)
	ctx := context.Background()
	task := &config.Task{
		Name: "helloworld",
		Cmd:  "echo {{ .JOB_ID }}",
	}
	jobConfig := &config.Job{}
	jobConfig.Tasks = append(jobConfig.Tasks, task)

	tasks := make(Tasks)
	for _, t := range jobConfig.Tasks {
		tasks[t.Name] = t
	}

	tctx := tasks.JSON()
	tctx["JOB_ID"] = "id"

	job := NewJob(ctx, "id", jobConfig)
	tr := NewTaskRunner(job, task, tctx)
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
	tr := NewTaskRunner(job, task, nil)

	tr.Run()

	<-job.ctx.Done()

	a.Equal(tr.fsm.Current(), "DONE")
	a.Equal(tr.State(), "DONE")
	t.Logf("task.output:%v", tr.Output())

}

func TestTaskRunnerHTTPPost(t *testing.T) {
	a := assert.Assert(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			t.Error(err)
			return
		}

		err = r.ParseMultipartForm(25)
		if err != nil {
			t.Error(err)
			return
		}

		dr, err := httputil.DumpRequest(r, true)
		t.Logf("req: %v", string(dr))
		t.Logf("p1: %v", r.PostFormValue("p1"))

		f, fh, err := r.FormFile("a.json")
		if err != nil {
			t.Error(err)
			return
		}
		var b bytes.Buffer
		_, err = b.ReadFrom(f)
		if err != nil {
			t.Error(err)
			return
		}

		t.Logf("%v", b.String())

		if b.Len() <= 0 {
			t.Error("file content is empty")
			return
		}

		a.Equal(r.Method, "POST")
		a.Equal(r.Form.Get("p1"), "1")
		a.Equal(r.Form.Get("p2"), "2")
		a.Equal(fh.Filename, "helloworld.json")

		if r.Method != "POST" {
			http.NotFound(w, r)
			return
		}

		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	var filePath string
	gopaths := filepath.SplitList(os.Getenv("GOPATH"))
	for _, gp := range gopaths {
		filePath = filepath.Join(gp, "src/github.com/go-loom/loom/test", "helloworld.json")
		if _, err := os.Stat(filePath); err == nil {
			break
		}
	}

	ctx := context.Background()
	task := &config.Task{
		Name: "http_post",
		HTTP: &config.HTTP{
			URL:    ts.URL,
			Method: "POST",
			Files: []*config.HTTPFile{
				&config.HTTPFile{
					Filename: "a.json",
					Path:     filePath,
				},
			},
			Data: map[string]string{
				"p1": "1",
				"p2": "2",
			},
		},
	}
	jobConfig := &config.Job{}
	jobConfig.Tasks = append(jobConfig.Tasks, task)

	job := NewJob(ctx, "id", jobConfig)
	tr := NewTaskRunner(job, task, nil)

	tr.Run()

	<-job.ctx.Done()

	a.Equal(tr.fsm.Current(), "DONE")
	a.Equal(tr.State(), "DONE")
	t.Logf("task.output:%v", tr.Output())

}

func TestNewTaskRunnerRetry(t *testing.T) {
	a := assert.Assert(t)
	ctx := context.Background()
	task := &config.Task{
		Name: "helloworld",
		Cmd:  "echo {{ .JOB_ID }} ; exit 1",
		Retry: config.Retry{
			Number:  3,
			Timeout: "0.1s",
		},
	}
	jobConfig := &config.Job{}
	jobConfig.Tasks = append(jobConfig.Tasks, task)

	tasks := make(Tasks)
	for _, t := range jobConfig.Tasks {
		tasks[t.Name] = t
	}

	tctx := tasks.JSON()
	tctx["JOB_ID"] = "id"

	job := NewJob(ctx, "id", jobConfig)
	tr := NewTaskRunner(job, task, tctx)
	tr.Run()

	<-job.ctx.Done()

	a.Equal(tr.fsm.Current(), "ERROR")
	a.Equal(tr.State(), "ERROR")
	t.Logf("task.output:%v", tr.Output())

}

func TestNewTaskRunnerRetryTimeout(t *testing.T) {
	a := assert.Assert(t)
	ctx := context.Background()
	task := &config.Task{
		Name: "helloworld",
		Cmd:  "echo {{ .JOB_ID }} ; sleep 1 ",
		Retry: config.Retry{
			Number:  3,
			Timeout: "0.1s",
		},
	}
	jobConfig := &config.Job{}
	jobConfig.Tasks = append(jobConfig.Tasks, task)

	tasks := make(Tasks)
	for _, t := range jobConfig.Tasks {
		tasks[t.Name] = t
	}

	tctx := tasks.JSON()
	tctx["JOB_ID"] = "id"

	job := NewJob(ctx, "id", jobConfig)
	tr := NewTaskRunner(job, task, tctx)
	tr.Run()

	<-job.ctx.Done()

	a.Equal(tr.fsm.Current(), "ERROR")
	a.Equal(tr.State(), "ERROR")
	t.Logf("task.output:%v", tr.Output())

}
