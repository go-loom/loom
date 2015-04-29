package config

import (
	"gopkg.in/yaml.v2"
	"testing"
)

func TestConfig1(t *testing.T) {
	y := []byte(`name: helloworld
version: 1.0.0
topic: test
tasks:
    - name: hello
      cmd: echo "hello world"
      when: JOB
      then: 
        state:
           hello: DONE
`)

	var worker Worker

	err := yaml.Unmarshal(y, &worker)
	if err != nil {
		t.Error(err)
	}

	task := worker.Tasks[0]

	if task.Name != "hello" {
		t.Error("worker task name should be hello.")
	}

	startStates := task.StartStates()

	if len(startStates) != 1 {
		t.Error("task start states is zero.")
	} else if startStates[0].Name != "JOB" {
		t.Error("task start state name should be JOB!")
	}

	endStates := task.EndStates()
	if len(endStates) != 1 {
		t.Error("task end states has one  state")
	} else if endStates[0].Name != "hello" {
		t.Errorf("task end state name should be hello %v", endStates[0].Name)
	}

}

func TestConfig2(t *testing.T) {
	y := []byte(`name: helloworld
version: 1.0.0
topic: test
tasks:
    - name: hello
      cmd: echo "hello world"
      when: 
        state:
          JOB: START
      then: 
        state:
           hello: DONE
`)

	var worker Worker

	err := yaml.Unmarshal(y, &worker)
	if err != nil {
		t.Error(err)
	}

	task := worker.Tasks[0]

	if task.Name != "hello" {
		t.Error("worker task name should be hello.")
	}

	startStates := task.StartStates()

	if len(startStates) != 1 {
		t.Error("task start states is zero.")
	} else if startStates[0].Name != "JOB" {
		t.Error("task start state name should be JOB!")
	}

	endStates := task.EndStates()
	if len(endStates) != 1 {
		t.Error("task end states has one  state")
	} else if endStates[0].Name != "hello" {
		t.Errorf("task end state name should be hello %v", endStates[0].Name)
	}

}
