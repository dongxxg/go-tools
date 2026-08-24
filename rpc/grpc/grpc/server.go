package main

import (
	"context"
	protoes "go-tools/pkg/rpc/grpc/myptoto"
	"google.golang.org/grpc"
	"log"
	"net"
)

type server struct {
	protoes.UnimplementedHelloServerServer
}

func (s *server) SayHello(ctx context.Context, in *protoes.HelloReq) (out *protoes.HelloRep, err error) {
	return &protoes.HelloRep{Msg: "Hello world!!!"}, nil
}

func (s *server) SayName(ctx context.Context, in *protoes.NameReq) (out *protoes.NameRep, err error) {
	return &protoes.NameRep{Msg: in.Name + " it is name"}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":10086")
	if err != nil {
		log.Println("network error", err)
	}
	srv := grpc.NewServer()
	protoes.RegisterHelloServerServer(srv, &server{})
	log.Printf("server listening at %+v\n", lis.Addr())
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
