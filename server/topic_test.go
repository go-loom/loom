package server

import (
	"fmt"
	"golang.org/x/net/context"
	"github.com/go-loom/loom/config"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func newTestStore() Store {
	err := os.Remove("/tmp/simple.boltdb")
	if err != nil {
		fmt.Printf("os.Remove err:%v\n", err)
	}

	store, _ := NewTopicStore("bolt", "/tmp", "simple")
	store.Open()
	return store
}

func newTestTopic() *Topic {
	store := newTestStore()

	quitC := make(chan struct{})
	ctx := context.WithValue(context.Background(), "quitC", quitC)

	topic := NewTopic(ctx, "simple", 1*time.Second, store)

	return topic
}

func TestTopicSimple(t *testing.T) {
	topic := newTestTopic()
	defer topic.store.Close()

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
	topic := newTestTopic()
	defer topic.store.Close()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var id MessageID
		m := NewMessage(id, nil)
		topic.PushMessage(m)
	}
}

func TestTopicJobRetry(t *testing.T) {
	topic := newTestTopic()
	defer topic.store.Close()

	retry := &config.Retry{
		Number:  2,
		Timeout: "1s",
	}
	job := &config.Job{
		Retry: retry,
	}

	var id MessageID
	copy(id[:], []byte("testid"))
	m := NewMessage(id, job)
	m.Created = m.Created.Add(-10 * time.Second)

	topic.msgBucket.Put(m)
	topic.pendingMsgBucket.Put(m)

	checkBuckets := func(expectedN int) {
		m2, err := topic.msgBucket.Get(id)
		if err != nil {
			t.Error(err)
			return
		}

		if m2.Job.Retry.NumRetry != expectedN {
			t.Errorf("don't retry checking %v", m2.Job.Retry.NumRetry)
		}

		m3, err := topic.pendingMsgBucket.Get(id)
		if m3.Job.Retry.NumRetry != expectedN {
			t.Errorf("don't retry checking %v", m3.Job.Retry.NumRetry)
		}
	}

	topic.checkRetryJobs()
	checkBuckets(1)

	m, err := topic.pendingMsgBucket.Get(id)
	if err != nil {
		t.Error(err)
	}

	ct := time.Now().Add(-10 * time.Second)
	m.Job.Retry.CheckedTime = &ct
	topic.pendingMsgBucket.Put(m)

	topic.checkRetryJobs()
	checkBuckets(2)

}

func TestTopicFinishReportUrl(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Error("method is not POST")
			return
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		t.Logf("body:%v", string(body))

	}))
	defer ts.Close()

	topic := newTestTopic()
	defer topic.store.Close()

	retry := &config.Retry{
		Number:  2,
		Timeout: "1s",
	}
	job := &config.Job{
		Retry:           retry,
		FinishReportURL: ts.URL,
	}

	var id MessageID
	copy(id[:], []byte("testid"))
	m := NewMessage(id, job)

	topic.msgBucket.Put(m)

	err := topic.FinishMessage(m.ID)
	if err != nil {
		t.Error(err)
		return
	}

}
