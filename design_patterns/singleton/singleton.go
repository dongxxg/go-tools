package singleton

import "sync"

type Foo struct {
	Name         string
	Data         []byte
}

var (
	instance       *Foo
	lock           sync.Mutex
)

func NewInstance1() *Foo {
	if instance == nil {
		lock.Lock()
		if instance == nil {
			instance = &Foo{
				Name: "name",
				Data: []byte{},
			}
		}
		lock.Unlock()
	}
	return instance
}

var once sync.Once

func NewInstance2() *Foo {
	once.Do(func() {
		instance = &Foo{
			Name: "name",
			Data: []byte{},
		}
	})
	return instance
}

var singletonInstance *Foo = &Foo{Name: "name", Data: []byte{}}
