package main

import (
	"log"
	"time"
)

func main() {
	go func() {
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-ticker.C:
				go func() {
					defer func() {
						if err := recover(); err != nil {
							log.Println(err)
						}
					}()
					proc()
				}()
			}
		}
	}()

	select {}
}

func proc() {
	panic("ok")
}
