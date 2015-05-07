package worker

import (
	"github.com/looplab/fsm"
	"testing"
)

func TestFSM(t *testing.T) {
	fsm := fsm.NewFSM(
		"init",
		fsm.Events{
			{Name: "run", Src: []string{"init"}, Dst: "process"},
			{Name: "success", Src: []string{"process"}, Dst: "done"},
			{Name: "error", Src: []string{"process"}, Dst: "done"},
		},
		fsm.Callbacks{},
	)

	if fsm.Current() != "init" {
		t.Error("fsm.Current() is not init")
	}

	if err := fsm.Event("run"); err != nil {
		t.Error(err)
	}

	if fsm.Current() != "process" {
		t.Error("fsm.Current() is not process")
	}

	if err := fsm.Event("success"); err != nil {
		t.Error(err)
	}

	if fsm.Current() != "done" {
		t.Error("fsm.Current() is not done")
	}

}
