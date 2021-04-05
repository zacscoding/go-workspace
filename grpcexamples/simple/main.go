package main

import (
	"context"
	pb "go-workspace/grpcexamples/simple/hello"
	"go-workspace/grpcexamples/simple/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"log"
	"time"
)

const (
	port = ":50051"
	addr = "localhost:50051"
)

func main() {
	// start gRPC server
	s, err := server.StartSimpleServer(port)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Stop()
	time.Sleep(time.Second)

	// start gRPC client
	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewHelloServiceClient(conn)

	callHello(c, "user", "coding")
	callHello(c, "user", "work")
	callHello(c, "user", "nothing")
	// Output
	//2021/04/05 18:00:27 [SERVER] received: firstName:"user"  lastName:"coding"
	//2021/04/05 18:00:27 [Client] received: greeting:"Hello zac-coding"
	//2021/04/05 18:00:27 [SERVER] received: firstName:"user"  lastName:"work"
	//2021/04/05 18:00:28 [Client] received status error. code:4, msg:context deadline exceeded
	//2021/04/05 18:00:28 [SERVER] received: firstName:"user"  lastName:"nothing"
	//2021/04/05 18:00:28 [Client] received status error. code:100, msg:unknown last name:nothing
}

func callHello(c pb.HelloServiceClient, firstName, lastName string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := c.Hello(ctx, &pb.HelloRequest{FirstName: firstName, LastName: lastName})
	if err != nil {
		if serr, ok := status.FromError(err); ok {
			log.Printf("[Client] received status error. code:%d, msg:%s", serr.Code(), serr.Message())
		} else {
			log.Printf("[Client] received error: %v", err)
		}
		return
	}
	log.Printf("[Client] received: %v", resp)
}
