package lru

import "container/list"

type LRU struct {
	ll             *list.List
	data           map[string]*list.Element
}

type entry struct {
	key   string
	value Value
}

// Value use Len to count how many bytes it takes
type Value interface {
	Len() int
}

func (l *LRU) Get(key string) (Value, bool) {
    if ele, ok := l.data[key]; ok {
    	l.ll.MoveToFront(ele)
    	v := ele.Value.(*entry)
    	return v.value, ok
	}
	return nil, false
}

func (l *LRU) removeOldest() {
    ele := l.ll.Back()
    if ele != nil {
    	l.ll.Remove(ele)
    	delete(l.data, ele.Value.(*entry).key)
	}
}

func (l *LRU) Put(key string, value Value) {
    if ele, ok := l.data[key]; ok {
    	l.ll.MoveToFront(ele)
    	kv := ele.Value.(*entry)
    	kv.value = value
	} else {
		l.ll.PushFront(&entry{value: value, key: key})
		l.data[key] = ele
	}
}
