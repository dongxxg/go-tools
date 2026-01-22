package rule

import "sync"

type FilterMsg struct {
	m map[string]interface{}
	sync.Mutex
}

func NewFilterMsg(m map[string]interface{}) *FilterMsg {
	return &FilterMsg{m: m}
}

func (f *FilterMsg) Get(k string) (interface{}, bool) {
	f.Lock()
	defer f.Unlock()
	if f.m == nil {
		f.m = make(map[string]interface{})
	}
	v, ok := f.m[k]
	return v, ok
}

func (f *FilterMsg) GetCondKey() (key, val string, exist bool) {
	//TODO implement me
	panic("implement me")
}
