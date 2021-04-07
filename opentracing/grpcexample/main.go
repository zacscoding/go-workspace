package main

import (
	"context"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	pb "go-workspace/opentracing/grpcexample/hello"
	"go-workspace/opentracing/grpcexample/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"time"
)

const (
	port = ":50051"
	addr = "localhost:50051"
)

// $ cd compose/jaeger
// $ docker-compose up
// $ go run main.go
// connect to jaeger-ui http://localhost:16686
func main() {
	// setup server tracer
	var (
		serverServiceName = "GRPC_Server"
		clientServiceName = "GRPC_Client"
	)
	serverTracer, serverCloser, err := newTracer(serverServiceName)
	if err != nil {
		log.Fatal(err)
	}
	defer serverCloser.Close()
	opentracing.SetGlobalTracer(serverTracer)

	// setup client tracer
	clientTracer, clientCloser, err := newTracer(clientServiceName)
	if err != nil {
		log.Fatal(err)
	}
	defer clientCloser.Close()

	// start gRPC server
	s, err := server.StartSimpleServer(port)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Stop()
	time.Sleep(time.Second)

	// start gRPC client
	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithUnaryInterceptor(
			otgrpc.OpenTracingClientInterceptor(clientTracer)),
		grpc.WithStreamInterceptor(
			otgrpc.OpenTracingStreamClientInterceptor(clientTracer)),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewHelloServiceClient(conn)

	callHello(c, "user", "coding", false)
	callHello(c, "user", "work", true)
	callHello(c, "user", "nothing", true)
}

func newTracer(serviceName string) (opentracing.Tracer, io.Closer, error) {
	tracerCfg := &jaegercfg.Configuration{
		ServiceName: serviceName,
		Disabled:    false,
		RPCMetrics:  false,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			BufferFlushInterval: time.Second,
			LocalAgentHostPort:  "localhost:6831",
		},
	}
	return tracerCfg.NewTracer()
}

func callHello(c pb.HelloServiceClient, firstName, lastName string, startSpan bool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if startSpan {
		var span opentracing.Span
		span, ctx = opentracing.StartSpanFromContext(ctx, "CallGRPC")
		defer span.Finish()
	}
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
