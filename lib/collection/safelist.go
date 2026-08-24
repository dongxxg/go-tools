package collection

import (
	"container/list"
	"sync"
)

type SafeList struct {
	sync.RWMutex
	L *list.List
}

func NewSafeList() *SafeList {
	return &SafeList{L: list.New()}
}

func (sl *SafeList) PushFront(v interface{}) *list.Element {
	sl.Lock()
	e := sl.L.PushFront(v)
	sl.Unlock()
	return e
}

func (sl *SafeList) PushFrontBatch(vs []interface{}) {
	sl.Lock()
	for _, item := range vs {
		sl.L.PushFront(item)
	}
	sl.Unlock()
}

func (sl *SafeList) PopBack() interface{} {
	sl.Lock()

	if elem := sl.L.Back(); elem != nil {
		item := sl.L.Remove(elem)
		sl.Unlock()
		return item
	}

	sl.Unlock()
	return nil
}

func (sl *SafeList) PopFront() interface{} {
	sl.Lock()

	if elem := sl.L.Front(); elem != nil {
		item := sl.L.Remove(elem)
		sl.Unlock()
		return item
	}

	sl.Unlock()
	return nil
}

func (sl *SafeList) PopBackBy(max int) []interface{} {
	sl.Lock()

	count := sl.len()
	if count == 0 {
		sl.Unlock()
		return []interface{}{}
	}

	if count > max {
		count = max
	}

	items := make([]interface{}, 0, count)
	for i := 0; i < count; i++ {
		item := sl.L.Remove(sl.L.Back())
		items = append(items, item)
	}

	sl.Unlock()
	return items
}

func (sl *SafeList) PopBackAll() []interface{} {
	sl.Lock()

	count := sl.len()
	if count == 0 {
		sl.Unlock()
		return []interface{}{}
	}

	items := make([]interface{}, 0, count)
	for i := 0; i < count; i++ {
		item := sl.L.Remove(sl.L.Back())
		items = append(items, item)
	}

	sl.Unlock()
	return items
}

func (sl *SafeList) Remove(e *list.Element) interface{} {
	sl.Lock()
	defer sl.Unlock()
	return sl.L.Remove(e)
}

func (sl *SafeList) RemoveAll() {
	sl.Lock()
	sl.L = list.New()
	sl.Unlock()
}

func (sl *SafeList) Front() interface{} {
	sl.RLock()

	if f := sl.L.Front(); f != nil {
		sl.RUnlock()
		return f.Value
	}

	sl.RUnlock()
	return nil
}

func (sl *SafeList) Back() interface{} {
	sl.RLock()

	if b := sl.L.Back(); b != nil {
		sl.RUnlock()
		return b.Value
	}

	sl.RUnlock()
	return nil
}

func (sl *SafeList) Len() int {
	sl.RLock()
	defer sl.RUnlock()
	return sl.len()
}

func (sl *SafeList) len() int {
	return sl.L.Len()
}

// SafeList with Limited Size
type SafeListLimited struct {
	maxSize int
	SL      *SafeList
}

func NewSafeListLimited(maxSize int) *SafeListLimited {
	return &SafeListLimited{SL: NewSafeList(), maxSize: maxSize}
}

func (sll *SafeListLimited) PopBack() interface{} {
	return sll.SL.PopBack()
}

func (sll *SafeListLimited) PopBackBy(max int) []interface{} {
	return sll.SL.PopBackBy(max)
}

func (sll *SafeListLimited) PopFront() interface{} {
	return sll.SL.PopFront()
}

func (sll *SafeListLimited) PushFront(v interface{}) bool {
	if sll.SL.Len() >= sll.maxSize {
		return false
	}

	sll.SL.PushFront(v)
	return true
}

func (sll *SafeListLimited) PushFrontBatch(vs []interface{}) bool {
	if sll.SL.Len() >= sll.maxSize {
		return false
	}

	sll.SL.PushFrontBatch(vs)
	return true
}

// PushFrontViolently 强制在Front上塞一个元素，把最后一个丢弃
func (sll *SafeListLimited) PushFrontViolently(v interface{}) bool {
	sll.SL.PushFront(v)
	if sll.SL.Len() > sll.maxSize {
		sll.SL.PopBack()
	}

	return true
}

func (sll *SafeListLimited) Remove(e *list.Element) interface{} {
	return sll.SL.Remove(e)
}

func (sll *SafeListLimited) RemoveAll() {
	sll.SL.RemoveAll()
}

func (sll *SafeListLimited) Front() interface{} {
	return sll.SL.Front()
}

func (sll *SafeListLimited) Back() interface{} {
	return sll.SL.Back()
}

func (sll *SafeListLimited) Len() int {
	return sll.SL.Len()
}

func (sll *SafeListLimited) PopBackAll() []interface{} {
	return sll.SL.PopBackAll()
}
