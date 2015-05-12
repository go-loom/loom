package worker

import (
	"bytes"
	"fmt"
	"github.com/seanpont/assert"
	"golang.org/x/net/context"
	"gopkg.in/loom.v1/worker/config"
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

	filePath := filepath.Join(
		os.Getenv("GOPATH"),
		"src/gopkg.in/loom.v1/test",
		"helloworld.json")

	ctx := context.Background()
	task := &config.Task{
		Name: "helloworld",
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
	tr := NewTaskRunner(job, task)

	tr.Run()

	<-job.ctx.Done()

	a.Equal(tr.fsm.Current(), "DONE")
	a.Equal(tr.State(), "DONE")
	t.Logf("task.output:%v", tr.Output())

}
