package main

import (
	"context"
	"log"
	"sync"
	"time"
)

func main() {
	var lock = &sync.Mutex{}
	pCtx := context.Background()
	ctx, cancel := context.WithCancel(pCtx)

	for i := 0; i < 10; i++ {
		go do(ctx, lock, i)
	}
	time.Sleep(2 * time.Second)
	cancel()
	time.Sleep(5 * time.Second)
	log.Println("exit...")
}

func do(ctx context.Context, lock *sync.Mutex, index int) {
	log.Printf("%d running...\n", index)
	time.Sleep(time.Duration(index) * time.Second)
	select {
	case <-ctx.Done():
		log.Printf("%d ctx done\n", index)
	default:
		defer func() {
			lock.Unlock()
			log.Printf("%d lock released\n", index)
		}()
		lock.Lock()
		log.Printf("%d acquire lock\n", index)
		log.Printf("%d done!!!\n", index)
	}
}
