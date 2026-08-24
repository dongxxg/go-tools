package fx

import (
	"time"
)

// Debounce 只有在不再收到trigger的interval时间后才会调用一次f()。
// 如果没有trigger，则永远不会执行。函数一旦被调用，goroutine一直存在
func Debounce(interval time.Duration, trigger chan interface{}, f func()) {
	var t bool
	for {
		select {
		case <-trigger:
			t = true
		case <-time.After(interval):
			func() {
				defer func() {
					t = false
				}()
				if t {
					f()
				}
			}()
		}
	}
}
