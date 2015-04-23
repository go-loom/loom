package server

import (
	"fmt"
)

type Queue interface {
	Cap() int64
	Len() int64
	Push(element interface{}) error
	Pop() interface{}
}

type LLNode struct {
	data interface{}
	next *LLNode
}

// Queue represents FIFO Queue data structure (linked list)

type LLQueue struct {
	head, tail  *LLNode
	size, items int64
}

func NewLLQueue(size ...int64) *LLQueue {
	queue := &LLQueue{}
	queue.size = -1
	if len(size) > 0 {
		queue.size = size[0]
	}

	return queue
}

// Return the Queue max capacity(size)
func (queue *LLQueue) Cap() int64 {
	return queue.size
}

// Return the Queue current length
func (queue *LLQueue) Len() int64 {
	return queue.items
}

// Push an element at the end of the queue
func (queue *LLQueue) Push(element interface{}) error {
	if queue.items == queue.size {
		return fmt.Errorf(
			"Can't push %v, Queue beyoud limits (%d)", element, queue.size)
	}

	n := &LLNode{data: element}
	if queue.tail == nil {
		queue.tail = n
		queue.head = n
	} else {
		queue.tail.next = n
		queue.tail = n
	}

	queue.items++
	return nil
}

// Pop an element from the head of the queue
func (queue *LLQueue) Pop() interface{} {
	if queue.items == 0 {
		return nil
	}

	head := queue.head
	data := head.data

	queue.head = head.next
	head = nil

	queue.items--

	return data
}
