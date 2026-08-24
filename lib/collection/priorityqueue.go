package collection

import "container/heap"

// 优先级队列实现，根据最小堆算法。队列中的元素需要时间Less方法
type Item interface {
	Less(other interface{}) bool
}

// 新建一个优先级队列
func NewPriorityQueue() *PriorityQueue {
	q := &PriorityQueue{s: new(sorter)}
	heap.Init(q.s)
	return q
}

type sorter []Item

// 实现heap.Interface: Push, Pop, Len, Less, Swap
func (s *sorter) Push(x interface{}) {
	*s = append(*s, x.(Item))
}

func (s *sorter) Pop() interface{} {
	n := len(*s)
	if n > 0 {
		x := (*s)[n-1]
		*s = (*s)[0 : n-1]
		return x
	}
	return nil
}

func (s *sorter) Len() int {
	return len(*s)
}

func (s *sorter) Less(i, j int) bool {
	return (*s)[i].Less((*s)[j])
}

func (s *sorter) Swap(i, j int) {
	(*s)[i], (*s)[j] = (*s)[j], (*s)[i]
}

// 优先级队列结构
type PriorityQueue struct {
	s *sorter
}

// 向堆插入元素
func (q *PriorityQueue) Push(x Item) {
	heap.Push(q.s, x)
}

// 从堆受弹出元素
func (q *PriorityQueue) Pop() Item {
	return heap.Pop(q.s).(Item)
}

// 获取堆首的元素，不弹出
func (q *PriorityQueue) Top() Item {
	if len(*q.s) > 0 {
		return (*q.s)[0]
	}
	return nil
}

// 更新堆上索引为i的元素，这时需要重建堆
func (q *PriorityQueue) Fix(x Item, i int) {
	(*q.s)[i] = x
	heap.Fix(q.s, i)
}

// 删除索引为i的元素
func (q *PriorityQueue) Remove(i int) Item {
	return heap.Remove(q.s, i).(Item)
}

// 返回堆中元素个数
func (q *PriorityQueue) Len() int {
	return q.s.Len()
}
