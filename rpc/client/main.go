package main

import (
    "fmt"
    rpc3 "go-tools/pkg/rpc"
    "log"
    "net/rpc"
)

type HelloServiceClient struct {
	*rpc.Client
}


func main() {
	client, err := rpc.Dial("tcp", "localhost:5673")
	if err != nil {
		log.Fatal("Dial tcp error:", err)
	}
	var reply string
	err = client.Call(rpc3.HelloServiceName+ "Hello", "dongxiang", &reply)
	if err != nil {
		log.Fatal("call error:", err)
	}
	fmt.Println(reply)
}


