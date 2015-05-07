package worker

import (
	"github.com/looplab/fsm"
	"github.com/seanpont/assert"
	"testing"
)

func TestFSM(t *testing.T) {
	assert := assert.Assert(t)
	fsm := fsm.NewFSM(
		"init",
		fsm.Events{
			{Name: "run", Src: []string{"init"}, Dst: "process"},
			{Name: "success", Src: []string{"process"}, Dst: "done"},
			{Name: "error", Src: []string{"process"}, Dst: "done"},
		},
		fsm.Callbacks{},
	)

	assert.Equal(fsm.Current(), "init")

	if err := fsm.Event("run"); err != nil {
		t.Error(err)
	}

	assert.Equal(fsm.Current(), "process")

	if err := fsm.Event("success"); err != nil {
		t.Error(err)
	}

	assert.Equal(fsm.Current(), "done")

}
