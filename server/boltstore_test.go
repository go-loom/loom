package server

import (
	"fmt"
	"github.com/koding/kite"
	"golang.org/x/net/context"
	"io"
	"os"
	"testing"
	"time"
)

func TestBoltStoreExpireMessage(t *testing.T) {

	s := NewBoltStore("/tmp/blotstore.test")
	if err := s.Open(); err != nil {
		t.Error(err)
	}
	defer func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	}()

	b := s.MessageBucket("test")
	if b == nil {
		t.Error(fmt.Errorf("bucket is nil"))
	}

	factory := &guidFactory{}
	id, err := factory.NewGUID(1)
	if err != nil {
		t.Error(err)
	}

	m := NewMessage(id.Hex(), nil)
	if err := b.Put(m); err != nil {
		t.Error(err)
	}

	m2, err := b.Get(m.ID)
	if err != nil {
		t.Error(err)
	}
	if m2 == nil {
		t.Error(fmt.Errorf("m2 is nil"))
	}

	b.(*BoltMessageBucket).ttl = 0 * time.Second

	m3, err := b.Get(m.ID)

	if err != io.EOF {
		t.Error(fmt.Errorf("m3 is io.EOF"))
	}

	if m3.ID == m.ID {
		t.Error(fmt.Errorf("m3.ID is same as m.ID"))
	}

}

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
