package concurrent

import (
	"github.com/orcaman/concurrent-map"
	"testing"
)

// @Author dx
// @Date 2025-07-26 22:46:00
// @Desc

type Person struct {
	Name string
	Age  int
}

func TestConcurrentMap(t *testing.T) {
	m := cmap.New()

	p := &Person{
		Name: "<UNK>",
		Age:  18,
	}
	m.Set("key", p)

	val, ok := m.Get("key")
	if ok {
		v := val.(*Person)
		t.Logf("get value:%#v", *v)
	}
}
