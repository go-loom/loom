package server

import (
	"container/list"
	"sync"
)

type Queue interface {
	Push(element interface{})
	Pop() interface{}
}

type LQueue struct {
	list *list.List
	mu   sync.Mutex
}

func NewLQueue() *LQueue {
	return &LQueue{list: list.New()}
}

func (q *LQueue) Push(elem interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.list.PushFront(elem)
}

func (q *LQueue) Pop() interface{} {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.list.Len() == 0 {
		return nil
	}

	e := q.list.Front()
	q.list.Remove(e)
	return e.Value
}

func (q *LQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.list.Len()
}

func (q *LQueue) List() []interface{} {
	ls := make([]interface{}, 0, q.list.Len())
	for e := q.list.Front(); e != nil; e = e.Next() {
		ls = append(ls, e.Value)
	}
	return ls
}
