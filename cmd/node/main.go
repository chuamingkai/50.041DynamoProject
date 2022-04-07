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

type UpdateNode struct {
	//Key    string `json:"key"`
	//PortNo uint64 `json:"portno"`
	NodeName uint64 `json:"nodename"`
}

// go run cmd/node/main.go -port PORT_NUMBER [-file FILE]
func main() {
	portPtr := flag.Int("port", 9000, "node port number")
	ringFile := flag.String("file", "", "file to import ring from")

	flag.Parse()

	portNumber := *portPtr
	file := *ringFile

	if err := config.LoadEnvFile(); err != nil {
		log.Fatalf("Error loading .env file: %v\n", err.Error())
	}

	var ring *consistenthash.Ring

	if file != "" {
		r, err := consistenthash.ImportRingFromFile(file)
		if err != nil {
			log.Fatal("Error importing ring from file ", file)
		}
		log.Printf("Ring imported: %s\n", r.Nodes.TraverseAndPrint())
		ring = r
	} else {
		ring = consistenthash.NewRing()
		ring.AddNode(strconv.Itoa(portNumber), uint64(portNumber)-3000)
		log.Printf("Ring imported: %s\n", ring.Nodes.TraverseAndPrint())

	}

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

	newnode := UpdateNode{NodeName: uint64(portNumber)}
	postBody, _ := json.Marshal(newnode)
	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post("http://localhost:8000/addNode", "application/json", responseBody)
	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	}

	defer resp.Body.Close()

	var newnodes []UpdateNode

	if err == nil {
		//fmt.Println(resp)
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
	}

	select {}
}
