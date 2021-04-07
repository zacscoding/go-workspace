package server

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	pb "go-workspace/opentracing/grpcexample/hello"
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

func (s *Server) Hello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	log.Printf("[SERVER] received: %v", req)
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		log.Println("[SERVER] Span is nil")
	} else {
		span.SetTag("req.firstName", req.FirstName)
		span.SetTag("req.lastName", req.LastName)
		if jspan, ok := span.(*jaeger.Span); ok {
			log.Printf("[SERVER] Operation Name: %s", jspan.OperationName())
			log.Printf("[SERVER] Span TraceID: %s", jspan.SpanContext().TraceID().String())
			log.Printf("[SERVER] Span ParentSanID: %s", jspan.SpanContext().ParentID().String())
			log.Printf("[SERVER] Span SpanID: %s", jspan.SpanContext().SpanID().String())
			for k, v := range jspan.Tags() {
				log.Printf("[SERVER] Tag key:%s, value:%v", k, v)
			}
		}
	}

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
	s := grpc.NewServer(
		grpc.UnaryInterceptor(
			otgrpc.OpenTracingServerInterceptor(opentracing.GlobalTracer())),
		grpc.StreamInterceptor(
			otgrpc.OpenTracingStreamServerInterceptor(opentracing.GlobalTracer()),
		),
	)
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
