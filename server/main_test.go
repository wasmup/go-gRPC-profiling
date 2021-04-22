package main

import (
	"context"
	"log"
	"net"
	"testing"

	pb "server/hello"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestSayHello(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatal("dial bufnet:", err)
	}
	defer conn.Close()
	client := pb.NewGreeterClient(conn)
	name := "name"
	resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		t.Fatal(err)
	}
	if resp.GetMessage() != "Hello name" {
		// log.Printf("Response: %+v", resp)
		t.Fatal(resp.GetMessage())
	}
}

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatal("Server exited with error", err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

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

	s := server{}
	for _, testCase := range testCases {
		req := &pb.HelloRequest{Name: testCase.name}
		resp, err := s.SayHello(context.Background(), req)
		if err != nil {
			log.Fatal("Server exited with error", err)
		}
		if resp.Message != testCase.want {
			t.Errorf("HelloText(%v)=%v, wanted %v", testCase.name, resp.Message, testCase.want)
		}
	}
}
