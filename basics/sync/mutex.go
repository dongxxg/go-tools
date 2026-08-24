package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

var (
	locker sync.RWMutex
	slice = make([]string, 0)
)

func addEle(s string) {
	locker.Lock()
	defer locker.Unlock()
	slice = append(slice, s)
	log.Println("add element", s)
	log.Println("slice==", slice)
}

func main() {
	for i := 0;i < 5; i++ {
		go addEle(fmt.Sprintf("ele_%d", i))
	}
	time.Sleep(5 * time.Second)
}