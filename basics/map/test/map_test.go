package test

import (
	"fmt"
	"log"
	"sync"
	"testing"
	"time"
)

type A struct {
	a string
}

func TestName(t *testing.T) {
	obj := &A{a: "1"}
	token := map[string]string{"a": obj.a}
	fmt.Println(token)

	obj.a = "2"
	fmt.Println(token)
}

func TestSyncMap(t *testing.T) {
	t.Helper()

	meetings := &sync.Map{}
	go store(meetings)
	time.Sleep(1 * time.Second)
	go store(meetings)

	time.Sleep(5 * time.Second)
}

func TestGet(t *testing.T) {
	m := map[int]bool{
		1: true,
	}

	if m[2] {
		t.Log("2222")
	} else {
		t.Log("3333")
	}
}

func store(meetings *sync.Map) {
	val := "struct{}{}"
	if _, ok := meetings.Load(1); !ok {
		if _, exist := meetings.LoadOrStore(1, val); !exist {
			storeVal, _ := meetings.Load(1)
			log.Println("store meetings ", storeVal)
		}
	}
}
