package server

import (
	"testing"
	"time"
)

func TestTopicSimple(t *testing.T) {

	store, _ := NewTopicStore("bolt", "/tmp", "simple")
	store.Open()
	defer store.Close()

	topic := NewTopic("simple", 1*time.Second, store)
	var id MessageID
	m := NewMessage(id, []byte{'1'})
	topic.PushMessage(m)

	m2, err := topic.store.GetMessage(m.ID)
	if err != nil {
		t.Error(err)
	}

	if m2 == nil {
		t.Errorf("data is nill,wants 1")
	}

	if string(m2.Value) != "1" {
		t.Errorf("data is not 1")
	}
}

func BenchmarkTopicPush(b *testing.B) {
	store, _ := NewTopicStore("bolt", "/tmp", "simple")
	store.Open()
	defer store.Close()

	topic := NewTopic("simple", 1*time.Second, store)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var id MessageID
		m := NewMessage(id, []byte{'1'})
		topic.PushMessage(m)
	}
}
