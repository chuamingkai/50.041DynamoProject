package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/chuamingkai/50.041DynamoProject/internal/nodes"
	consistenthash "github.com/chuamingkai/50.041DynamoProject/pkg/consistenthashing"
	pb "github.com/chuamingkai/50.041DynamoProject/pkg/internalcomm"
	"google.golang.org/grpc"
)

// go run cmd/node/main.go -port PORT_NUMBER
func main() {
	portPtr := flag.Int("port", 9000, "node port number")
	flag.Parse()

	portNumber := *portPtr
	ring := consistenthash.NewRing()

	// TODO: Node discovery
	ring.AddNode(strconv.Itoa(portNumber), uint64(portNumber))

	server := nodes.NewNodeServer(int64(portNumber), ring)
	serverPtr := server.CreateServer()

	// Create gRPC listener
	grpcAddress := fmt.Sprintf("localhost:%v", *portPtr - 3000)
	grpcListener, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		log.Fatalf("Failed to listen to grpc address %v: %v", grpcAddress, err)
	}

	// TODO: Run a grpc client as well
	// Register to gRPC
	grpcServer := grpc.NewServer()
	pb.RegisterReplicationServer(grpcServer, server)

	go func () {
		log.Println("Server running at port", portNumber)
		if err := serverPtr.ListenAndServe(); err != nil {
			log.Fatalf("Terminating node %v due to error: %v\n", portNumber, err.Error())
		}
	}()
	

	// Serve gRPC
	go func() {
		log.Printf("Serving gRPC server at %v\n", grpcAddress)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("Failed to serve grpc address %v: %v", grpcAddress, err)
		}
	}()

	select{}
}
