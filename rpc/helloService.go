package rpc

import "net/rpc"

const HelloServiceName = "HelloService"

type HelloServiceInterface interface {
	Hello(request string, reply *string) error
}

func RegistryHelloService(inter HelloServiceInterface) error {
	return rpc.RegisterName(HelloServiceName, inter)
}