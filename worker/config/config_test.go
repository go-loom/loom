package config

import (
	"testing"
)

func TestConfigRead(t *testing.T) {
	y := []byte(`name: helloworld
version: 1.0.0
topic: test
tasks:
	- name: hello
	  cmd: echo "hello world ID:{{ .job.ID }} "
	  when: JOB
	  then: 
	  	state:
			hello: DONE
`)

	worker, err := Read(y)
	if err != nil {
		t.Error(err)
		return
	}

	if worker.Name != "helloworld" {
		t.Error("wrong worker name")
	}

	for _, task := range worker.Tasks {
		ss := task.StartStates()
		ed := task.EndStates()
		t.Log(ss)
		t.Log(ed)
	}

}
