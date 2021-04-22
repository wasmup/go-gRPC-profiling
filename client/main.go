package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	pb "client/hello"

	"google.golang.org/grpc"
)

func main() {
	{
		f, err := os.Create("cpu.out")
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	name := "world"
	var r *pb.HelloReply
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	t0 := time.Now()
	count := 0
loop1:
	for {
		r, err = c.SayHello(ctx, &pb.HelloRequest{Name: name})
		if err != nil {
			fmt.Println(err)
			break loop1
		}
		count++
		// select {
		// case <-ctx.Done():
		// 	break loop1
		// default:
		// }
	}
	fmt.Println("d=", time.Since(t0))
	fmt.Printf("Greeting: %s\n", r.GetMessage())
	fmt.Println("count=", count)
	{
		f, err := os.Create("mem.out")
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}
