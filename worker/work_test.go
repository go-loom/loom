package worker

import (
	"golang.org/x/net/context"
	"gopkg.in/loom.v1/worker/config"
	"strings"
	"testing"
)

func readYamlConfig() *config.Worker {
	y := []byte(`name: helloworld
version: 1.0.0
topic: test
tasks:
	- name: hello
	  cmd: echo {{ .ID }}
	  when: JOB
	  then: 
	  	state:
			hello: DONE
	- name: world
	  cmd: echo world
	  when: hello
	  then: JOB
`)

	worker, err := config.Read(y)
	if err != nil {
		return nil
	}

	return worker
}

func TestNewWork(t *testing.T) {
	job := NewJob("hello", map[string]interface{}{"URL": "http://example.com"})
	workerConfig := readYamlConfig()
	if workerConfig == nil {
		t.Error("worker config failed")
		return
	}

	ctx := context.Background()
	work := NewWork(ctx, job, workerConfig)
	work.Run()
	<-work.Done()

	if work.Err() != nil {
		t.Error(work.Err())
	}

	if work.Running() {
		t.Error("The work is processing")
	}

	for _, task := range work.Tasks() {
		if !task.Ok() {
			t.Errorf("task: %v err: %v output: %v", task.Name(), task.Err(), task.Output())
		} else if task.Name() == "hello" {
			if !strings.Contains(task.Output(), "hello") {
				t.Errorf("task: %v err: %v output: %v", task.Name(), task.Err(), task.Output())
			}
		} else if task.Name() == "world" {
			if !strings.Contains(task.Output(), "world") {
				t.Errorf("task: %v err: %v output: %v", task.Name(), task.Err(), task.Output())
			}
		}
	}
}
