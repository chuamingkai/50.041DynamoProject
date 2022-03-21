package nodes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/chuamingkai/50.041DynamoProject/internal/bolt"
	"github.com/chuamingkai/50.041DynamoProject/internal/models"
	consistenthash "github.com/chuamingkai/50.041DynamoProject/pkg/consistenthashing"
	"github.com/chuamingkai/50.041DynamoProject/pkg/vectorclock"

	"github.com/gorilla/mux"
)

var db bolt.DB
var a *consistenthash.Ring
var nodeId int64
type UpdateNode struct {
	NodeName string `json:"nodename"`
}


func createSingleEntry(w http.ResponseWriter, r *http.Request) {

	/* TODO: To add check for if node handles key*/
	nodename := strings.Trim(r.Host, "localhost:")

	reqBody, _ := ioutil.ReadAll(r.Body)
	var newEntry models.Object
	json.Unmarshal(reqBody, &newEntry)

	number, err := strconv.ParseUint(nodename, 10, 64)
	if err != nil {
		log.Fatalf("Error parsing node name: %s", err)
	}
	if !a.IsNodeResponsibleForKey(newEntry.Key, number) {
		http.Error(w, "Node is not responsible for key!", 400)
		return
	}

	if newEntry.VC != nil {
		newEntry.VC = vectorclock.UpdateRecv(nodename, newEntry.VC)
	} else {
		newEntry.VC = vectorclock.UpdateRecv(nodename, map[string]uint64{nodename: 0})
	}

	// Insert test value into bucket
	err = db.Put("testBucket", newEntry)
	if err != nil {
		log.Fatalf("Error inserting into bucket: %s", err)
	}

	/* TODO: To add ignore if new updated vector clock is outdated?*/

	// TODO: Create context to get replicas
	// clientDeadline := time.Now().Add(time.Second * 5)
	// ctxRep, cancelRep := context.WithDeadline(context.TODO(), clientDeadline)
	// defer cancelRep()

	// TODO: Send replicas to other nodes
	// request := &rp.PutRepRequest{
	// 	Key: []byte(newEntry.Key),
	// 	Data: []byte(newEntry.Value), // Not sure if this is correct
		
	// }
	// preferenceList := a.GetPreferenceList(newEntry.Key)
	// for _, virtualNode := range preferenceList {
	// 	go func (virtualNodeId uint64)  {
	// 		if virtualNodeId== uint64(nodeId) {
	// 			// Insert test value into bucket
	// 			err = db.Put("testBucket", newEntry)
	// 			if err != nil {
	// 				log.Fatalf("Error inserting into bucket: %s", err)
	// 			}
	// 		} else {
	// 			// Send replica to other nodes
	// 			var peer rp.ReplicationClient
	// 			// var response *rp.PutRepResponse
	
	// 			peer.PutReplica(ctxRep, request)
	// 		}
	// 	}(virtualNode.NodeId)
	// }

	// TODO: Check if replica has been successfully received by servers in preference list

	/*Edit to return a list of objects instead*/
	json.NewEncoder(w).Encode(newEntry)
}

func returnSingleEntry(w http.ResponseWriter, r *http.Request) {
	// Check if node handles key
	nodename := strings.Trim(r.Host, "localhost:")
	number, _ := strconv.ParseUint(nodename, 10, 64)

	vars := mux.Vars(r)
	key := vars["ic"]

	if !a.IsNodeResponsibleForKey(key, number) {
		http.Error(w, "Node is not responsible for key!", 400)
		return
	}
	// id, err := strconv.Atoi(nodename)
	fmt.Println(a.GetPreferenceList(key))

	// Read from bucket
	value, err := db.Get("testBucket", key)
	if err != nil {
		log.Fatal("Error getting from bucket:", err)
	}

	// TODO: Get replicas from servers in preference list
	// preferenceList := a.GetPreferenceList(key)
	// replicas := make([]rp.GetRepResponse, 100) // TODO: Figure out how many replicas needed
	// request := rp.GetRepRequest{
	// 	Key: []byte(key),
	// }

	// for _, virtualNode := range preferenceList {
	// 	go func(virtualNodeId int64) {
	// 		var dataReplica *rp.GetRepResponse
	// 		var err error

	// 		if virtualNodeId != nodeId {
	// 			var peer rp.ReplicationClient
	// 			// var response *rp.PutRepResponse
	
	// 			peer.PutReplica(ctxRep, request)
	// 		}
	// 	}(int64(virtualNode.NodeId))
	// }

	// Send list if have conflicts, else send data

	if value.Key != "" {
		// TODO: Encode list of conflicting replicas
		json.NewEncoder(w).Encode(value)
	} else {
		http.Error(w, "Value not found!", 404)
	}

}

func updateAddNode(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)

	var newnode UpdateNode
	json.Unmarshal(reqBody, &newnode)
	exist := false
	number, err := strconv.ParseUint(newnode.NodeName, 10, 64)
	if err == nil {
		for _, n := range a.NodeMap {
			if n.NodeId == number {
				exist = true
			}
		}
		if !exist {
			a.AddNode(newnode.NodeName, number)
			fmt.Println(a.Nodes.TraverseAndPrint())
		} else {
			http.Error(w, "Node already exists!", 503)
		}
	}
	json.NewEncoder(w).Encode(newnode)
}

func updateDelNode(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)

	var newnode UpdateNode
	json.Unmarshal(reqBody, &newnode)
	exist := false
	number, err := strconv.ParseUint(newnode.NodeName, 10, 64)
	if err == nil {
		for _, n := range a.NodeMap {
			if n.NodeId == number {
				exist = true
			}
		}
		if exist {
			a.RemoveNode(newnode.NodeName)
			fmt.Println(a.Nodes.TraverseAndPrint())
		} else {
			http.Error(w, "Node doesn't exists!", 503)
		}
	}
	json.NewEncoder(w).Encode(newnode)
}

func CreateServer(port int, newRing *consistenthash.Ring) *http.Server {
	fmt.Println("Setting up node at port", port)
	nodeId = int64(port)
	var err error

	// Open database
	db, err = bolt.ConnectDB(port)
	if err != nil {
		log.Fatal("Error connecting to Bolt database:", err)
	}
	// defer db.DB.Close()

	// Add ring
	a = newRing

	// Create bucket
	err = db.CreateBucket("testBucket")
	if err != nil {
		log.Fatal("Error creating bucket:", err)
	}

	/*creates a new instance of a mux router*/
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/addnode", updateAddNode).Methods("POST")
	myRouter.HandleFunc("/delnode", updateDelNode).Methods("POST")

	/*write new entry*/
	myRouter.HandleFunc("/data", createSingleEntry).Methods("POST")

	/*return single entry*/
	myRouter.HandleFunc("/data/{ic}", returnSingleEntry)

	server := http.Server{
		Addr:    fmt.Sprintf(":%v", port), //:{port}
		Handler: myRouter,
	}
	return &server
}