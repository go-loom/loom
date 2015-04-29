package config

import (
	"gopkg.in/yaml.v2"
	"testing"
)

func TestYamlSimple(t *testing.T) {
	y := []byte(
		`name: test
version: 1.0.1
tasks:
    - name: task1
      cmd: python helloworld.py 
      when: JOB
    - name: task2
      cmd: python helloworld.py 
`)

	var worker Worker
	err := yaml.Unmarshal(y, &worker)
	if err != nil {
		t.Error(err)
	}

	if len(worker.Tasks) != 2 {
		t.Error("worker tasks not exist")
	}

	t.Log(worker.Tasks)

}
