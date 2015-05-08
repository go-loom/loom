package worker

import (
	"github.com/looplab/fsm"
	a "github.com/seanpont/assert"
	"testing"
)

func TestFSM(t *testing.T) {
	assert := a.Assert(t)

	fsm := fsm.NewFSM(
		"init",
		fsm.Events{
			{Name: "run", Src: []string{"init"}, Dst: "process"},
			{Name: "success", Src: []string{"process"}, Dst: "success"},
			{Name: "cancel", Src: []string{"init", "process"}, Dst: "cancel"},
			{Name: "error", Src: []string{"process"}, Dst: "error"},
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
	assert.Equal(fsm.Current(), "success")

}
