package main

import (
	"fmt"
	"sync"
)

// way one

type SafeMap1 struct {
	Map map[string]string
	sync.RWMutex
}

func NewSafeMap1() *SafeMap1 {
	sm := new(SafeMap1)
	sm.Map = make(map[string]string)
	return sm
}

func (s *SafeMap1) store(key, value string) {
	defer s.Unlock()
	s.Lock()
	s.Map[key] = value
}

func (s *SafeMap1) load(key string) string {
	defer s.RUnlock()
	s.RLock()
	value := s.Map[key]
	return value
}

// way two

const N = 16

type SafeMap2 struct {
	Maps  [N]map[string]string
	locks [N]sync.RWMutex
}

func NewSafeMap2() *SafeMap2 {
	sm := new(SafeMap2)
	for i := 0; i < N; i++ {
		sm.Maps[i] = make(map[string]string)
	}
	return sm
}

func (m *SafeMap2) store(key, value string) {
	index := m.hash(key) % N
	m.locks[index].Lock()
	m.Maps[index][key] = value
	m.locks[index].Unlock()
}

func (m *SafeMap2) load(key string) string {
	index := m.hash(key) % N
	m.locks[index].RLock()
	value := m.Maps[index][key]
	m.locks[index].RUnlock()
	return value
}

func (m *SafeMap2) hash(s string) int {
	h := 0
	for i := 0; i < len(s); i++ {
		h = 31*h + int(s[i])
	}
	return h
}

func main() {
	safeMap := NewSafeMap2()

	var wg sync.WaitGroup

	// 启动多个goroutine进行写操作
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			safeMap.store(fmt.Sprintf("name%d", i), fmt.Sprintf("John%d", i))
		}(i)
	}

	wg.Wait()

	// 启动多个goroutine进行读操作
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			fmt.Println(safeMap.load(fmt.Sprintf("name%d", i)))
		}(i)
	}

	wg.Wait()

}
