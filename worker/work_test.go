package worker

import (
	"golang.org/x/net/context"
	"gopkg.in/loom.v1/worker/config"
	"strings"
	"testing"
)

func testRunWork(y []byte, job Job) (*Work, error) {
	workerConfig, err := config.Read(y)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	work := NewWork(ctx, job, workerConfig)
	work.Run()
	<-work.Done()
	return work, nil
}

func TestWorkTasks(t *testing.T) {
	job := NewJob("hello", map[string]interface{}{"URL": "http://example.com"})
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

	work, err := testRunWork(y, job)
	if err != nil {
		t.Error(err)
		if work == nil {
			return
		}
	}

	if work.Err() != nil {
		t.Error(work.Err())
	}

	if work.Running() {
		t.Error("The work is processing")
	}

	for _, task := range work.tasks {
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

func TestWorkConcurrentTasks(t *testing.T) {
	job := NewJob("hello", map[string]interface{}{"URL": "http://example.com"})
	y := []byte(`name: helloworld
version: 1.0.0
topic: test
tasks:
	- name: hello
	  cmd: echo {{ .ID }}
	  when: JOB
	  then: JOB
	- name: world
	  cmd: echo world
	  when: JOB
	  then: JOB
`)

	work, err := testRunWork(y, job)
	if err != nil {
		t.Error(err)
		if work == nil {
			return
		}
	}

	if work.Err() != nil {
		t.Error(work.Err())
	}

	if work.Running() {
		t.Error("The work is processing")
	}

	for _, task := range work.tasks {
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
