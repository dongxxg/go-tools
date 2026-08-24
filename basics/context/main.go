package main

import (
	"fmt"
	"math/rand"
	"time"
)

func GenerateRequestID() string {
	// 生成一个随机的请求标识
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%d", rand.Intn(1000))
}

func main() {
	fmt.Println(GenerateRequestID())
}
