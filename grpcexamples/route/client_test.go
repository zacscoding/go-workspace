package route

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go-workspace/grpcexamples/person"
	pb "go-workspace/grpcexamples/route/person"
	"go-workspace/grpcexamples/route/server"
	"google.golang.org/grpc"
	"io"
	"math/rand"
	"testing"
	"time"
)

const (
	serverPort = ":50051"
	clientAddr = "localhost:50051"
)

type ClientSuite struct {
	suite.Suite
	serv       *server.Server
	clientConn *grpc.ClientConn
	cli        pb.PersonRouteClient
}

func (s *ClientSuite) TestGetPersonWithMonitoring() {
	conns := make(chan *grpc.ClientConn, 50)

	for i := 0; i < 50; i++ {
		conn, err := grpc.Dial(clientAddr, grpc.WithInsecure(), grpc.WithBlock())
		s.NoError(err)
		conns <- conn
	}

	time.Sleep(time.Minute)
}

func (s *ClientSuite) TestGetPerson() {
	// when
	resp, err := s.cli.GetPerson(context.Background(), &pb.PersonQuery{
		Id: 1,
	})
	// then
	s.NoError(err)
	s.EqualValues(1, resp.Id)
	s.Equal("user1", resp.Name)
	s.EqualValues(12, resp.Age)
}

func (s *ClientSuite) TestListPerson() {
	// given
	userName := "user-" + uuid.New().String()
	writeStream, err := s.cli.SavePerson(context.Background())
	s.NoError(err)
	for i := 0; i < 10; i++ {
		s.NoError(writeStream.Send(&pb.PersonRequest{
			Name: userName,
			Age:  rand.Int31n(50),
		}))
	}
	_, err = writeStream.CloseAndRecv()
	s.NoError(err)

	// when
	stream, err := s.cli.ListPerson(context.Background(), &pb.PersonQuery{
		Name: userName,
	})
	s.NoError(err)
	count := 0
	for {
		p, err := stream.Recv()
		if err == io.EOF {
			break
		}
		s.NoError(err)
		count++
		s.Equal(userName, p.Name)
	}
	s.Equal(10, count)
}

func (s *ClientSuite) TestGetPersonChat() {
	// given
	var userNames []string
	writeStream, err := s.cli.SavePerson(context.Background())
	s.NoError(err)
	for i := 0; i < 3; i++ {
		userName := "user-" + uuid.New().String()
		for i := 0; i < 5; i++ {
			s.NoError(writeStream.Send(&pb.PersonRequest{
				Name: userName,
				Age:  rand.Int31n(50),
			}))
		}
		userNames = append(userNames, userName)
	}
	_, err = writeStream.CloseAndRecv()
	s.NoError(err)

	stream, err := s.cli.GetPersonChat(context.Background())
	s.NoError(err)

	waitc := make(chan struct{})
	received := 0
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				s.NoError(err)
				s.Contains(userNames, in.Name)
				received++
			}
		}
	}()

	for _, userName := range userNames {
		err := stream.Send(&pb.PersonQuery{Name: userName})
		s.NoError(err)
	}
	stream.CloseSend()
	<-waitc
	s.Greater(received, 0)
}

func (s *ClientSuite) TestSavePerson() {
	// given
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := s.cli.SavePerson(ctx)
	s.NoError(err)
	personList := []*person.Person{
		{Name: "auser", Age: 10},   // fail
		{Name: "user", Age: 15},    // success
		{Name: "another", Age: 29}, // fail
		{Name: "user", Age: 15},    // success
	}

	// when
	for _, p := range personList {
		err := stream.Send(&pb.PersonRequest{
			Name: p.Name,
			Age:  p.Age,
		})
		s.NoError(err)
	}
	reply, err := stream.CloseAndRecv()
	s.NoError(err)

	// then
	s.EqualValues(len(personList), reply.Trial)
	s.EqualValues(2, reply.Success)
	s.EqualValues(2, reply.Fail)
	s.Greater(reply.Elapsed, int64(0))
}

func TestRunClientSuite(t *testing.T) {
	suite.Run(t, new(ClientSuite))
}

func (s *ClientSuite) SetupSuite() {
	// setup gRPC server
	serv, err := server.StartServer(serverPort)
	s.NoError(err)
	s.serv = serv

	// setup gRPC client
	conn, err := grpc.Dial(clientAddr, grpc.WithInsecure(), grpc.WithBlock())
	s.NoError(err)
	s.clientConn = conn
	s.cli = pb.NewPersonRouteClient(conn)
}

func (s *ClientSuite) TearDownTest() {
	s.clientConn.Close()
	s.serv.Stop()
}
