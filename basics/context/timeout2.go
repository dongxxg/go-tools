package main

// 启动 2 个 groutine 2 秒后取消，第一个协程 1 秒 执行完，第二个协程 3 秒执行完。
// 思路：采用 ctx, _ := context.WithTimeout(context.Background(), time.Second*2)实现 2s 取消。
// 协程执行完后通过 channel 通知，是否超时。

import (
	"context"
	"fmt"
	"time"
)

func f1(in chan struct{}) {
	time.Sleep(1 * time.Second)
	in <- struct{}{}
}

func f2(in chan struct{}) {
	time.Sleep(3 * time.Second)
	in <- struct{}{}
}

func main() {
	ch1 := make(chan struct{})
	ch2 := make(chan struct{})
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	go func() {
		go f1(ch1)
		select {
		case <-ctx.Done():
			fmt.Println("f1 timeout")
			break
		case <-ch1:
			fmt.Println("f1 done")
		}
	}()
	go func() {
		go f2(ch2)
		select {
		case <-ctx.Done():
			fmt.Println("f2 timeout")
			break
		case <-ch2:
			fmt.Println("f2 done")
		}
	}()
	time.Sleep(time.Second * 5)
}
