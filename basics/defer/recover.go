package main

import (
	"errors"
	"fmt"
	"log"
	"runtime"
)

type test struct {
	A *string
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			panic(r)
		}
	}()

	t := new(test)
	panic(errors.New("test error"))
	fmt.Println(*t.A)
}

func runWithRecovery() {
	defer func() {
		if r := recover(); r != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Printf("cron: panic running job: %v\n%s", r, buf)
		}
	}()
	// do something... panic
}
