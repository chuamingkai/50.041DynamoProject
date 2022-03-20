package main

import (
	"fmt"

	"github.com/chuamingkai/50.041DynamoProject/internal/nodes"
	consistenthash "github.com/chuamingkai/50.041DynamoProject/pkg/consistenthashing"
)

func main() {
	name := "9000"
	a := consistenthash.NewRing()

	a.AddNode(name, 9000)

	server := nodes.CreateServer(9000, a)
	fmt.Println(server.ListenAndServe())

}
