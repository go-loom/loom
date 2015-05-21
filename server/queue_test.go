package server

import (
	"testing"
)

func TestQueueSimple(t *testing.T) {
	q := NewLQueue()

	q.Push(1)
	if l := q.Len(); l != 1 {
		t.Errorf("q.Len() = %d,want %d", l, 1)

	}

	x := q.Pop()

	if l := q.Len(); l != 0 {
		t.Errorf("q.Len() = %d,want %d", l, 0)

	}

	if x.(int) != 1 {
		t.Errorf("q.Pop() = %v,want %v", x, 1)
	}

}

func BenchmarkQueuePush(b *testing.B) {
	q := NewLQueue()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		q.Push(i)
	}
}

func BenchmarkQueuePushAndPop(b *testing.B) {
	q := NewLQueue()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		q.Push(i)
	}

	for i := 0; i < b.N; i++ {
		q.Pop()
	}

}
