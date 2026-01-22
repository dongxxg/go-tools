package client

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestGoRouteing(t *testing.T) {
	_, cf := context.WithCancel(context.TODO())

	go func() {
		go func() {
			for {
				fmt.Println("here")
				time.Sleep(time.Second)
			}
		}()
		time.Sleep(time.Second * 5)
		cf()
	}()

	select {}
}
