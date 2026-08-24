package main

import (
	"context"
	"fmt"
	"os"
	"time"
)

func searchTarget(ctx context.Context, data []int, target int, ch chan bool) {
	for _, v := range data {
		select {
		case <-ctx.Done():
			fmt.Fprintf(os.Stdout, "Task cancelded! \n")
			return
		default:
			fmt.Fprintf(os.Stdout, "v: %d \n", v)
			time.Sleep(time.Millisecond * 1500)
			if target == v {
				ch <- true
				return
			}
		}
	}
}

func main() {
	timer := time.NewTimer(5 * time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	resultChannel := make(chan bool, 1)
	target := 345
	numList := []int{1, 2, 3, 10, 999, 8, 345, 7, 98, 33, 66, 77, 88, 68, 96}
	numLen := len(numList)
	size := 3
	for i := 0; i < numLen; i += size {
		end := i + size
		if end >= numLen {
			end = numLen - 1
		}
		go searchTarget(ctx, numList[i:i+3], target, resultChannel)
	}
	select {
	case <-timer.C:
		fmt.Fprintln(os.Stderr, "Timeout! Not Found")
		cancel()
	case <-resultChannel:
		fmt.Fprintf(os.Stdout, "Found it!\n")
		cancel()
	}
}
