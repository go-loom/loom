package server

import (
	"sync"
	"testing"
	"time"
)

func TestTopicSimple(t *testing.T) {

	store, _ := NewTopicStore("bolt", "/tmp", "simple")
	store.Open()
	defer store.Close()

	topic := NewTopic("simple", 1*time.Second, store)
	var id MessageID
	m := NewMessage(id, 1)
	topic.PushMessage(m)
	data := topic.PopMessage()
	if data == nil {
		t.Errorf("data is nill,wants 1")
	}

	if data.Value.(int) != 1 {
		t.Errorf("data is not 1")
	}
}

func _TestTopicPopBlock(t *testing.T) {
	var wg sync.WaitGroup

	store, _ := NewTopicStore("bolt", "/tmp", "simple")
	store.Open()
	defer store.Close()

	topic := NewTopic("simple", 1*time.Second, store)
	go func() {
		var id MessageID
		m := NewMessage(id, 1)
		topic.PushMessage(m)
	}()

	go func() {
		wg.Add(1)
		m := topic.PopMessage()
		if m.Value.(int) != 1 {
			t.Errorf("PopMessage should be blocked %v", m)
		}
		wg.Done()
	}()
	wg.Wait()

}

func BenchmarkTopicPush(b *testing.B) {
	store, _ := NewTopicStore("bolt", "/tmp", "simple")
	store.Open()
	defer store.Close()

	topic := NewTopic("simple", 1*time.Second, store)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var id MessageID
		m := NewMessage(id, i)
		topic.PushMessage(m)
	}
}
