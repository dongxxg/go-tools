package concurrent

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	LockedFlag   int32 = 1
	UnlockedFlag int32 = 0
)

type Mutex struct {
	in sync.Mutex
}

func NewMutex() *Mutex {
	return &Mutex{}
}

func (m *Mutex) Lock() {
	m.in.Lock()
}

func (m *Mutex) Unlock() {
	m.in.Unlock()
}

func (m *Mutex) TryLock() bool {
	if atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(&m.in)), UnlockedFlag, LockedFlag) {
		return true
	}
	return false
}

//func (m *Mutex) IsLocked() bool {
//	if atomic.LoadInt32((*int32)(unsafe.Pointer(&m.in))) == LockedFlag {
//		return true
//	}
//	return false
//}
