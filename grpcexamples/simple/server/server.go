package server

import (
	"context"
	"fmt"
	pb "go-workspace/grpcexamples/simple/hello"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"strings"
	"time"
)

type Server struct {
	pb.UnimplementedHelloServiceServer
	s *grpc.Server
}

func (s *Server) Stop() {
	if s.s != nil {
		s.s.Stop()
	}
}

func (s *Server) Hello(_ context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	log.Printf("[SERVER] received: %v", req)
	switch strings.ToLower(req.LastName) {
	case "coding":
		return &pb.HelloResponse{Greeting: fmt.Sprintf("Hello %s-%s", req.FirstName, req.LastName)}, nil
	case "work":
		time.Sleep(time.Second * 2)
		return &pb.HelloResponse{Greeting: fmt.Sprintf("Hello %s-%s", req.FirstName, req.LastName)}, nil
	default:
		return nil, status.Errorf(100, "unknown last name:%s", req.LastName)
	}
}

func StartSimpleServer(addr string) (*Server, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	s := grpc.NewServer()
	server := &Server{
		s: s,
	}
	pb.RegisterHelloServiceServer(s, server)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	return server, nil
}
