package main

import (
	"fmt"
	"time"
)

func now() {
	location := time.Now().Location()
	fmt.Println(location) //Local
	now := time.Now().In(location)
	fmt.Println(now) // now
	fmt.Println(now.Format("2006-01-02_15:04:05"))

	fmt.Println(time.Now().UnixMilli())
}

func main() {
	now()
	timeFormat(1733205917012)
	timeFormat(1733476036415)
}

func currentTime() {
	s := time.Now().UnixMilli() + 2592000000
	fmt.Println(time.Now().UnixMilli())
	fmt.Println(s)
	fmt.Println(time.Duration(2592000000) * time.Second)

}

func timeFormat(t int64) {
	T := time.UnixMilli(t)
	fmt.Println(T.Format("2006-01-02_15:04:05"))
}
