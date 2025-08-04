package concurrent

import (
	"golang.org/x/sync/singleflight"
	"sync"
	"testing"
	"time"
)

func TestSF(t *testing.T) {
	var group = singleflight.Group{}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err, _ := group.Do("key", func() (interface{}, error) {
				time.Sleep(1 * time.Second)
				t.Logf("[1_%d] do exec success", i)
				return "value", nil
			})
			if err != nil {
				t.Errorf("error:%v\n", err)
				return
			}
			t.Logf("[1_%d]get result success", i)
		}()
	}

	wg.Wait()
	t.Log("================================================")

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err, _ := group.Do("key", func() (interface{}, error) {
				time.Sleep(1 * time.Second)
				t.Logf("[2_%d] do exec success", i)
				return "value", nil
			})
			if err != nil {
				t.Errorf("error:%v\n", err)
				return
			}
			t.Logf("[2_%d]get result success", i)
		}(i)
	}
	wg.Wait()
}
