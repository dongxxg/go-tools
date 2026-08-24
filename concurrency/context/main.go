package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func work(ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()
	for {
		select {
		default:
			fmt.Println("hello")
			case <-ctx.Done():
				return ctx.Err()
		}
	}
}

func testContext() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	var wg sync.WaitGroup
	for i := 0;i < 10;i++ {
		wg.Add(1)
		go work(ctx, &wg)
	}
	time.Sleep(time.Second)
	cancel()
	wg.Wait()
}

func generateNatural(ctx context.Context) chan int {
	ch := make(chan int)
    go func() {
    	for i := 2; ;i++ {
			select {
			case <- ctx.Done():
				return
    		case ch <- i:
			}
		}
    }()
	return ch
}

func primeFilter(ctx context.Context, in <-chan int, prime int) chan int {
    out := make(chan int)
    go func() {
    	for {
    		if i := <-in; i%prime != 0 {
				select {
    			case <- ctx.Done():
    				return
    				case out <- i:
				}
			}
		}
	}()
	return out
}

func testPrimeFilter() {
	ctx, cancel := context.WithCancel(context.Background())
	ch := generateNatural(ctx)
	for i := 0;i < 100;i++ {
		prime := <- ch
		fmt.Printf("%v: %v\n", i + 1, prime)
		ch = primeFilter(ctx, ch, prime)
	}
	cancel()
}

func main() {
    testPrimeFilter()
}
