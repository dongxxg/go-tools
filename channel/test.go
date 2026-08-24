package main

import (
	"fmt"
	"math/rand"
	"sync"
)

func main() {
	ch := make(chan int)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			ch <- rand.Intn(5)
		}
		close(ch)
	}()

	go func() {
		defer wg.Done()
		for v := range ch {
			fmt.Println(v)
		}
	}()
	wg.Wait()
}
