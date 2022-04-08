package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/chuamingkai/50.041DynamoProject/config"
	"github.com/chuamingkai/50.041DynamoProject/internal/nodes"
	consistenthash "github.com/chuamingkai/50.041DynamoProject/pkg/consistenthashing"
	pb "github.com/chuamingkai/50.041DynamoProject/pkg/internalcomm"
	"google.golang.org/grpc"
)

var (
	portNumber int
	ringFile string
)

type UpdateNode struct {
	//Key    string `json:"key"`
	//PortNo uint64 `json:"portno"`
	NodeName uint64 `json:"nodename"`
}

// Import ring
func importRing() *consistenthash.Ring {
	var ring *consistenthash.Ring
	if ringFile != "" {
		r, err := consistenthash.ImportRingFromFile(ringFile)
		if err != nil {
			log.Fatal("Error importing ring from file ", ringFile)
		}
		log.Printf("Ring imported: %s\n", r.Nodes.TraverseAndPrint())
		ring = r
	} else {
		ring = consistenthash.NewRing()
		ring.AddNode(strconv.Itoa(portNumber), uint64(portNumber)-3000)
		log.Printf("Ring imported: %s\n", ring.Nodes.TraverseAndPrint())

	}
	return ring
}

// Register node with node manager
func registerNode(ring *consistenthash.Ring) {
	newnode := UpdateNode{NodeName: uint64(portNumber)}
	postBody, _ := json.Marshal(newnode)
	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post(config.REGISTRATION_URL, "application/json", responseBody)
	if err != nil {
		log.Fatalln("Error registering node:", err.Error())
		return
	}
	defer resp.Body.Close()

	var newnodes []UpdateNode
	jsonDataFromHttp, err1 := ioutil.ReadAll(resp.Body)
	if err1 == nil {
		json.Unmarshal([]byte(jsonDataFromHttp), &newnodes)
		if len(newnodes) > 0 {
			for _, n := range newnodes {

				ring.AddNode(strconv.FormatUint(n.NodeName, 10), n.NodeName-3000)
				log.Printf("Notified of the existence of Node %v", n.NodeName)

			}
			log.Println(ring.Nodes.TraverseAndPrint())
		}
	}
	log.Println("Finished registration with node manager")
}

// go run cmd/node/main.go -port PORT_NUMBER [-file FILE]
func main() {
	portPtr := flag.Int("port", 9000, "node port number")
	ringFilePtr := flag.String("file", "", "file to import ring from")

	flag.Parse()

	portNumber = *portPtr
	ringFile = *ringFilePtr

	if err := config.LoadEnvFile(); err != nil {
		log.Fatalf("Error loading .env file: %v\n", err.Error())
	}

	ring := importRing()

	server := nodes.NewNodeServer(int64(portNumber), int64(portNumber)-3000, ring)
	serverPtr, err := server.RunNodeServer()
	if err != nil {
		log.Fatalf("Failed to start node server: %v", err.Error())
	}

	// Create gRPC listener
	grpcAddress := fmt.Sprintf("localhost:%v", *portPtr-3000)
	grpcListener, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		log.Fatalf("Failed to listen to grpc address %v: %v", grpcAddress, err)
	}

	// Register to gRPC
	grpcServer := grpc.NewServer()
	pb.RegisterReplicationServer(grpcServer, server)

	// Serve HTTP
	go func() {
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

	registerNode(ring)

	select {}
}
