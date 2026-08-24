package main

import (
	"fmt"
	"time"
)

func f1(in chan int) {
	fmt.Println(<-in)
}

func test1() {
	out := make(chan int, 2)
	out <- 2
	go f1(out)
	time.Sleep(3 * time.Second)
}

func testDeadlock() {
	out := make(chan int)
	out <- 2
	go f1(out)
	time.Sleep(3 * time.Second)
}

func DeadlockCorrect() {
	out := make(chan int)
	go func() {
		out <- 2
	}()
	go f1(out)
	select {}
}

func main() {
	DeadlockCorrect()
}
