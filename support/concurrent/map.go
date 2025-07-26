package concurrent

import "sync"

// @Author dx
// @Date 2025-07-26 22:44:00
// @Desc

type SafeMap struct {
	mu sync.Mutex
	m  map[string]interface{}
}

func (s *SafeMap) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] = value
}

func (s *SafeMap) Get(key string) (interface{}, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	val, ok := s.m[key]
	return val, ok
}
