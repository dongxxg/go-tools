package redis

import (
	"net"
)

const (
	Address = "127.0.0.1:6379"
	Network = "tcp"
	)


func Conn(network, address string) (net.Conn, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return conn, nil
}



