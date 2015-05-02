package config

import (
	"bytes"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var (
	ConfigNoTasks = errors.New("config has no tasks")
)

func ReadFile(path string) (*Worker, error) {
	yamlc, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	worker, werr := Read(yamlc)
	if werr != nil {
		return nil, werr
	}
	return worker, nil

}

func Read(yamlcontent []byte) (*Worker, error) {

	//Replace tab to 2 space (yaml format don't accept tab)
	content := bytes.Replace(yamlcontent, []byte{'\t'}, []byte{' ', ' ', ' ', ' '}, -1)

	var worker Worker

	err := yaml.Unmarshal(content, &worker)
	if err != nil {
		return nil, err
	}

	if len(worker.Tasks) <= 0 {
		return nil, ConfigNoTasks
	}

	for _, task := range worker.Tasks {
		ss := task.StartStates()
		if len(ss) <= 0 {
			return nil, fmt.Errorf("Task %s has not start state", task.Name)
		}
		es := task.EndStates()
		if len(es) <= 0 {
			return nil, fmt.Errorf("Task %s has not end state", task.Name)
		}
	}

	return &worker, nil
}
