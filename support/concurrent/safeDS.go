package concurrent

import "sync"

type SafeStringSlice struct {
	s []string
	m sync.Mutex
}

func (s *SafeStringSlice) Append(str string) {
	s.m.Lock()
	defer s.m.Unlock()
	if s.s == nil {
		s.s = make([]string, 0)
	}
	s.s = append(s.s, str)
}

func (s *SafeStringSlice) GetSlice() []string {
	return s.s
}
