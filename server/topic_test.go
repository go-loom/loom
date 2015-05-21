package server

import (
	"golang.org/x/net/context"
	"testing"
	"time"
)

func TestTopicSimple(t *testing.T) {

	store, _ := NewTopicStore("bolt", "/tmp", "simple")
	store.Open()
	defer store.Close()

	quitC := make(chan struct{})
	ctx := context.WithValue(context.Background(), "quitC", quitC)

	topic := NewTopic(ctx, "simple", 1*time.Second, store)
	var id MessageID
	m := NewMessage(id, nil)
	topic.PushMessage(m)

	m2, err := topic.msgBucket.Get(m.ID)
	if err != nil {
		t.Error(err)
	}

	if m2 == nil {
		t.Errorf("data is nill,wants 1")
	}

}

func BenchmarkTopicPush(b *testing.B) {
	store, _ := NewTopicStore("bolt", "/tmp", "simple")
	store.Open()
	defer store.Close()

	ctx := context.Background()
	topic := NewTopic(ctx, "simple", 1*time.Second, store)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var id MessageID
		m := NewMessage(id, nil)
		topic.PushMessage(m)
	}
}
