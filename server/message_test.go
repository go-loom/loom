package server

import (
	"testing"
)

func TestMessageID(t *testing.T) {
	b := NewBroker("/tmp/test")
	b.Init()
	m := NewMessage(b.NewID(), 1)
	if len(m.ID) <= 0 {
		t.Errorf("messageID is %v", m.ID)
	}

	t.Logf("%s", m.ID)
}
