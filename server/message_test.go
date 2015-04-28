package server

import (
	"github.com/koding/kite"
	"testing"
)

func TestMessageID(t *testing.T) {
	k := kite.New("test", "0.0.1")
	b := NewBroker("/tmp/test", k)
	b.Init()
	m := NewMessage(b.NewID(), []byte{'1'})
	if len(m.ID) <= 0 {
		t.Errorf("messageID is %v", m.ID)
	}

	t.Logf("%s", m.ID)
}
