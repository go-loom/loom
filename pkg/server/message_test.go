package server

import (
	"context"
	"testing"
)

func TestMessageID(t *testing.T) {
	ctx := context.Background()
	b := NewBroker(ctx, "/tmp/test")
	b.Init()
	m := NewMessage(b.NewID(), nil)
	if len(m.ID) <= 0 {
		t.Errorf("messageID is %v", m.ID)
	}

	t.Logf("%s", m.ID)
}
