package socket

import (
	"net"
)

func (s *Server) Dial(addr string) net.Conn {

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		s.ctx.Fatal(err)
		return nil
	}
	defer conn.Close()

	return conn
}
