package concurrent

import "testing"

func TestQueue(t *testing.T) {
	q := Queue{}
	println(q.Empty())
	println(q.Pop())
	q.Push(1)
	println(q.Pop())
	println(q.Empty())
}
