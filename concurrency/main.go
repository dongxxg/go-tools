package main

import (
	"fmt"
	"sync"
	"time"
)

var ch = make(chan int, 3)
var work = []func() {
	func() {println("1"); time.Sleep(1 * time.Second)},
	func() {println("2"); time.Sleep(1 * time.Second)},
	func() {println("3"); time.Sleep(1 * time.Second)},
	func() {println("4"); time.Sleep(1 * time.Second)},
	func() {println("5"); time.Sleep(1 * time.Second)},
}

func testChan() {
	for _, w := range work {
		go func(w func()) {
			ch <- 1
			w()
			<- ch
		}(w)
	}
	select {}
}


func testChan2() {
	done := make(chan int, 10)
	for i := 0;i < cap(done); i++ {
		go func() {
			println("hello world")
			done <- 1
		}()
	}

	for i := 0; i < cap(done);i++ {
		<- done
	}
}

// testWaitGroup use Chan-->use WaitGroup
func testWaitGroup() {
	var wg sync.WaitGroup
	for i := 0;i < 10; i++ {
		wg.Add(1)
		go func() {
			println("hello world")
			wg.Done()
		}()
	}
	wg.Wait()
}

// testConPub consumer and producer
func testConsumerProducer()  {
	out := make(chan int, 64)
    go produce(3, out)
    go produce(5, out)
	go consume(out)
	time.Sleep(1*time.Second)
}

func produce(factor int, ch chan <-int) {
    for i := 0;;i++{
        ch <- i * factor
	}
}

func consume(ch <-chan int) {
    for v := range ch {
    	println(v)
	}
}

func worker() {
	ticker := time.Tick(1 * time.Second)
	for {
		select {
		case <- ticker:
			fmt.Println("run 1s worker")
		}
	}
}

var (
	count int
	lock sync.RWMutex
)

func read() {
	for {
		lock.RLock()
		fmt.Println("read ", count)
        lock.RUnlock()
	}
}

func write() {
	for {
		lock.Lock()
		count++
		lock.Unlock()
	}
}

func main() {
	for i := 0;i < 10;i++{
		go read()
	}

	for i := 0;i < 5;i++ {
		go write()
	}
	fmt.Scanln()
}
