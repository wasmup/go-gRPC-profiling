package main

import (
	"time"

	pb "client/hello"

	"context"
	"log"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestSayHello2(t *testing.T) {
	testCases := []struct{ name, want string }{
		{
			name: "world",
			want: "Hello world",
		},
		{
			name: "everyone",
			want: "Hello everyone",
		},
	}

	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			c := pb.NewGreeterClient(conn)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			r, err := c.SayHello(ctx, &pb.HelloRequest{Name: testCase.name})
			if err != nil {
				log.Fatalf("could not greet: %v", err)
			}
			if r.GetMessage() != testCase.want {
				t.Errorf("HelloText(%v)=%v, wanted %v", testCase.name, r.GetMessage(), testCase.want)
			}
		})
	}
}

type mockServer struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *mockServer) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func (s *mockServer) SayHelloAgain(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello again " + in.GetName()}, nil
}

func (s *mockServer) Add(ctx context.Context, in *pb.AddRequest) (*pb.AddReply, error) {
	return &pb.AddReply{Sum: in.GetA() + in.GetB()}, nil
}

func dialer() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	pb.RegisterGreeterServer(server, &mockServer{})

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}
