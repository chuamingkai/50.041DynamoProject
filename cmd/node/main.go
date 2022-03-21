package main

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/chuamingkai/50.041DynamoProject/internal/nodes"
	consistenthash "github.com/chuamingkai/50.041DynamoProject/pkg/consistenthashing"
)

// go run cmd/node/main.go -port PORT_NUMBER
func main() {
	portPtr := flag.Int("port", 9000, "node port number")
	flag.Parse()

	portNumber := *portPtr
	ring := consistenthash.NewRing()

	ring.AddNode(strconv.Itoa(portNumber), uint64(portNumber))

	server := nodes.CreateServer(portNumber, ring)
	fmt.Println(server.ListenAndServe())

}
