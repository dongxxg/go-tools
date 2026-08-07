package collection

import "sync"

// A QueueV2 is a FIFO queue.
type QueueV2[T any] struct {
	lock     sync.Mutex
	elements []T
	size     int
	head     int
	tail     int
	count    int
}

// NewQueueV2 returns a QueueV2 object.
func NewQueueV2[T any](size int) *QueueV2[T] {
	return &QueueV2[T]{
		elements: make([]T, size),
		size:     size,
	}
}

// Empty checks if q is empty.
func (q *QueueV2[T]) Empty() bool {
	q.lock.Lock()
	empty := q.count == 0
	q.lock.Unlock()

	return empty
}

// Put puts element into q at the last position.
func (q *QueueV2[T]) Put(element T) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.head == q.tail && q.count > 0 {
		nodes := make([]T, len(q.elements)+q.size)
		copy(nodes, q.elements[q.head:])
		copy(nodes[len(q.elements)-q.head:], q.elements[:q.head])
		q.head = 0
		q.tail = len(q.elements)
		q.elements = nodes
	}

	q.elements[q.tail] = element
	q.tail = (q.tail + 1) % len(q.elements)
	q.count++
}

// Take takes the first element out of q if not empty.
func (q *QueueV2[T]) Take() (T, bool) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.count == 0 {
		var t T
		return t, false
	}

	element := q.elements[q.head]
	q.head = (q.head + 1) % len(q.elements)
	q.count--

	return element, true
}
