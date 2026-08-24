package main

import (
	"context"
	"fmt"
	"time"
)

func doSomething(ctx context.Context) {
	time.Sleep(5 * time.Second)
	if ctx.Err() != nil {
		fmt.Println("doing something is canceled")
		return
	}
	fmt.Println("doing something is end")
}

func main() {
	timeout := 3 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	fmt.Println("running...")
	for i := 0; i < 3; i++ {
		go doSomething(ctx)
	}
	select {
	case <-ctx.Done():
		fmt.Println(ctx.Err())
		fmt.Println("end")
	}
}
