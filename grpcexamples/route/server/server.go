package server

import (
	"context"
	"go-workspace/grpcexamples/person"
	pb "go-workspace/grpcexamples/route/person"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"
)

type Server struct {
	pb.UnimplementedPersonRouteServer
	s *grpc.Server
	r *person.Repository
}

func StartServer(addr string) (*Server, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	s := grpc.NewServer()
	server := &Server{
		s: s,
		r: person.NewRepository(),
	}
	pb.RegisterPersonRouteServer(s, server)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	return server, nil
}

func (s *Server) Stop() {
	if s.s != nil {
		s.s.Stop()
	}
}

// GetPerson tests simple rpc
func (s *Server) GetPerson(_ context.Context, query *pb.PersonQuery) (*pb.PersonResponse, error) {
	log.Printf("[Server] Called GetPerson(). query: %v", query)
	result, err := s.r.FindByID(query.Id)
	if err != nil {
		if err == person.ErrNotFound {
			return nil, status.Errorf(404, "not found person. id:%d", query.Id)
		} else {
			return nil, status.Error(codes.Unknown, err.Error())
		}
	}
	return &pb.PersonResponse{
		Id:   result.ID,
		Name: result.Name,
		Age:  result.Age,
	}, nil
}

// ListPerson tests server-to-client streaming RPC
func (s *Server) ListPerson(query *pb.PersonQuery, stream pb.PersonRoute_ListPersonServer) error {
	log.Printf("[Server] Called ListPerson(). query: %v", query)
	results := s.r.FindAllByName(query.Name)
	for _, result := range results {
		if err := stream.Send(&pb.PersonResponse{
			Id:   result.ID,
			Name: result.Name,
			Age:  result.Age,
		}); err != nil {
			return err
		}
	}
	return nil
}

// SavePerson tests client-to-server streaming RPC
func (s *Server) SavePerson(stream pb.PersonRoute_SavePersonServer) error {
	var (
		start   = time.Now()
		trial   = int32(0)
		success = int32(0)
		fail    = int32(0)
	)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			endTime := time.Now().Sub(start)
			return stream.SendAndClose(&pb.PersonSaveSummary{
				Trial:   trial,
				Success: success,
				Fail:    fail,
				Elapsed: int64(endTime.Seconds()),
			})
		}
		if err != nil {
			return err
		}
		trial++
		log.Printf("[Server] SavePerson(). trial:%d, req:%v", trial, req)
		if strings.HasPrefix(req.Name, "a") {
			fail++
		} else {
			success++
			s.r.Save(&person.Person{
				Name: req.Name,
				Age:  req.Age,
			})
		}
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	}
}

// GetPersonChat tests bidirectional streaming RPC
func (s *Server) GetPersonChat(stream pb.PersonRoute_GetPersonChatServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		result, err := s.r.FindByID(in.Id)
		if err != nil {
			return err
		}
		if err := stream.Send(&pb.PersonResponse{
			Id:   result.ID,
			Name: result.Name,
			Age:  result.Age,
		}); err != nil {
			return err
		}
	}
}
