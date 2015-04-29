package config

import (
	"gopkg.in/yaml.v2"
	"testing"
)

func TestConfig(t *testing.T) {
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
	t.Log(task)

	if task.Name != "hello" {
		t.Error("worker task name should be hello.")
	}

	t.Logf("task.when: %v", task.When)
	t.Logf("task.then: %v", task.Then)

	if s := task.StartState(); s == nil || s.Name != "JOB" {
		t.Error("worker when attribute should be JOB.")
	}

}
