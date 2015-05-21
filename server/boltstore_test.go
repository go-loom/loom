package server

import (
	"github.com/koding/kite"
	"golang.org/x/net/context"
	"os"
	"testing"
)

func TestBoltStorePut(t *testing.T) {
	ctx := context.Background()
	k := kite.New("test", "0.0.1")
	b := NewBroker(ctx, "/tmp/test", k)
	b.Init()

	s, err := NewTopicStore("bolt", b.DBPath, "test")
	if err != nil {
		t.Error(err)
	}
	err = s.Open()
	if err != nil {
		t.Error(err)
	}

	defer func() {
		s.Close()
		os.Remove(s.(*BoltStore).Path)
	}()

	m := NewMessage(b.NewID(), nil)
	mb := s.MessageBucket("test")

	err = mb.Put(m)
	if err != nil {
		t.Error(err)
	}

	var m1 *Message
	m1, err = mb.Get(m.ID)
	if err != nil {
		t.Error(err)
		return
	}

	if m1 == nil {
		t.Error("m1 is nil!")
		return
	}

	if m1.ID != m.ID {
		t.Errorf("ID %s,wants %s", m1.ID, m.ID)
	}

}
