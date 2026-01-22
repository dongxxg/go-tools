/**
  @Author: dongxx
  @Since: 2025/10/13 15:39
  @Desc: //TODO
**/

package main

import (
	"log"
	"time"
)

var ch1 = make(chan struct{}, 2)
var ch2 = make(chan struct{}, 2)

func test2() {
	for {
		select {
		case _, ok := <-ch2:
			if ok {
				log.Println("receive ch2 msg")
				ch1 <- struct{}{}
			} else {
				log.Println("close ch2")
				return
			}
		}
	}
}

func test3() {
	for {
		select {
		case _, ok := <-ch1:
			if ok {
				log.Println("receive ch1 msg")
			} else {
				log.Println("close ch1")
				return
			}
		}
	}
}

func main() {
	go test2()
	go test3()

	go func() {
		time.Sleep(2 * time.Second)
		close(ch1)
		close(ch2)
	}()

	go func() {
		for i := 0; i < 3; i++ {
			ch1 <- struct{}{}
			log.Println("send msg to ch1")
		}
	}()

	time.Sleep(5 * time.Second)
}
