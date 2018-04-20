package server

import (
	"github.com/koding/kite"
	"github.com/go-loom/loom/config"
	"sync"
	"testing"
)

func TestDispatcherRunning(t *testing.T) {
	k := kite.New("test", "0.0.1")
	client := k.NewClient("http://localhost")
	client.ID = "test1"

	topicName := "test"

	w := NewConnectedWorker("1", topicName, 1, client)

	dispatcher := NewDispatcher(topicName, nil)
	dispatcher.AddWorker(w)

	if dispatcher.running != true {
		t.Errorf("dispatcher has 1 workers but it not running!")
	}

	dispatcher.RemoveWorker(w)

	if dispatcher.running == true {
		t.Errorf("dispatcher should be not running.")
	}
}

func TestDispatcherSendMessage(t *testing.T) {
	k := kite.New("test", "0.0.1")
	client := k.NewClient("http://localhost")
	client.ID = "test1"

	topicName := "test"

	w := NewConnectedWorker("1", topicName, 1, client)

	dispatcher := NewDispatcher(topicName, nil)
	dispatcher.AddWorker(w)

	factory := &guidFactory{}
	id, _ := factory.NewGUID(int64(0))

	var wg sync.WaitGroup

	go func() {
		wg.Add(1)
		msg := NewMessage(MessageID(id.Hex()), &config.Job{})
		dispatcher.msgPopChan <- msg
		wg.Done()
	}()

	wg.Wait()
	msg := <-dispatcher.msgPushChan
	if msg == nil {
		t.Errorf("dispatcher can't requeue a msg of failure to send")
	}

}
