package main

import (
	"log"
	"net"
	"net/rpc"
)

type HelloService struct {}

func (h *HelloService) Hello(request string, reply *string) error {
    *reply = "hello" + request
    return nil
}

func main() {
	rpc.RegisterName("HelloService", new(HelloService))
	listener, err := net.Listen("tcp", ":5673")
	if err != nil {
		log.Fatal("listenTCP error:", err)
	}
	conn, err := listener.Accept()
	if err != nil {
		log.Fatal("Accept error:", err)
	}
	rpc.ServeConn(conn)
}
