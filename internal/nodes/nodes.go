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

// TODO: Name this smth else
type nodesServer struct {
	boltDB *bolt.DB
	ring *consistenthash.Ring
	nodeId int64
	repServer *replicaServer
}

// TODO: Move the structs below into pkg/models
type UpdateNode struct {
	NodeName string `json:"nodename"`
}

type PutRequestBody struct {
	BucketName string `json:"bucketName"`
	Object models.Object `json:"object"`
}

type GetRequestBody struct {
	BucketName string `json:"bucketName"`
	Key string `json:"key"`
}


func (s *nodesServer) doPut(w http.ResponseWriter, r *http.Request) {
	nodename := strings.Trim(r.Host, "localhost:")

	reqBodyBytes, _ := ioutil.ReadAll(r.Body)
	var reqBody PutRequestBody
	json.Unmarshal(reqBodyBytes, &reqBody)

	// Check if bucket name is provided
	if reqBody.BucketName == "" {
		http.Error(w, "Missing bucket name", http.StatusBadRequest)
		return
	}

	// Check if bucket exists
	if !s.boltDB.BucketExists(reqBody.BucketName) {
		http.Error(w, "Bucket does not exist", http.StatusBadRequest)
		return
	}

	number, err := strconv.ParseUint(nodename, 10, 64)
	if err != nil {
		log.Fatalf("Error parsing node name: %s", err)
	}

	// Check if node handles key
	if !s.ring.IsNodeResponsibleForKey(reqBody.Object.Key, number) {
		http.Error(w, "Node is not responsible for key!", http.StatusBadRequest)
		return
	}

	if reqBody.Object.VC != nil {
		reqBody.Object.VC = vectorclock.UpdateRecv(nodename, reqBody.Object.VC)
	} else {
		reqBody.Object.VC = vectorclock.UpdateRecv(nodename, map[string]uint64{nodename: 0})
	}

	// Insert test value into bucket
	err = s.boltDB.Put(reqBody.BucketName, reqBody.Object)
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
	json.NewEncoder(w).Encode(reqBody.Object)
}

func (s *nodesServer) doGet(w http.ResponseWriter, r *http.Request) {
	reqBodyBytes, _ := ioutil.ReadAll(r.Body)
	var reqBody GetRequestBody
	json.Unmarshal(reqBodyBytes, &reqBody)

	// Check if bucket name and/or key is provided
	if reqBody.BucketName == "" || reqBody.Key == "" {
		http.Error(w, "Missing key/bucket name", http.StatusBadRequest)
		return
	}

	// Check if bucket exists
	if !s.boltDB.BucketExists(reqBody.BucketName) {
		http.Error(w, "Bucket does not exist", http.StatusBadRequest)
		return
	}

	// Check if node handles key
	nodename := strings.Trim(r.Host, "localhost:")
	number, _ := strconv.ParseUint(nodename, 10, 64)

	if !s.ring.IsNodeResponsibleForKey(reqBody.Key, number) {
		http.Error(w, "Node is not responsible for key!", http.StatusBadRequest)
		return
	}
	// id, err := strconv.Atoi(nodename)
	// fmt.Println(s.ring.GetPreferenceList(key))

	// Read from bucket
	log.Println("Reading from bucket", reqBody.BucketName)
	value, err := s.boltDB.Get(reqBody.BucketName, reqBody.Key)
	if err != nil {
		// log.Fatal("Error getting from bucket:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
		http.Error(w, "Value not found!", http.StatusNotFound)
	}

}

func (s *nodesServer) updateAddNode(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)

	var newnode UpdateNode
	json.Unmarshal(reqBody, &newnode)
	exist := false
	number, err := strconv.ParseUint(newnode.NodeName, 10, 64)
	if err == nil {
		for _, n := range s.ring.NodeMap {
			if n.NodeId == number {
				exist = true
			}
		}
		if !exist {
			s.ring.AddNode(newnode.NodeName, number)
			fmt.Println(s.ring.Nodes.TraverseAndPrint())
		} else {
			http.Error(w, "Node already exists!", http.StatusServiceUnavailable)
		}
	}
	json.NewEncoder(w).Encode(newnode)
}

func (s *nodesServer) updateDelNode(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)

	var newnode UpdateNode
	json.Unmarshal(reqBody, &newnode)
	exist := false
	number, err := strconv.ParseUint(newnode.NodeName, 10, 64)
	if err == nil {
		for _, n := range s.ring.NodeMap {
			if n.NodeId == number {
				exist = true
			}
		}
		if exist {
			s.ring.RemoveNode(newnode.NodeName)
			fmt.Println(s.ring.Nodes.TraverseAndPrint())
		} else {
			http.Error(w, "Node doesn't exists!", http.StatusServiceUnavailable)
		}
	}
	json.NewEncoder(w).Encode(newnode)
}

func (s *nodesServer) doCreateBucket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var newBucketName string
	var ok bool
	if newBucketName, ok = vars["bucketName"]; !ok {
		http.Error(w, "Missing bucket name", http.StatusBadRequest)
	}

	// TODO: Create bucket on replicas as well
	err := s.boltDB.CreateBucket(newBucketName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
}

func (s *nodesServer) CreateServer() *http.Server {
	fmt.Println("Setting up node at port", s.nodeId)

	/*creates a new instance of a mux router*/
	myRouter := mux.NewRouter().StrictSlash(true)

	// Node handling
	myRouter.HandleFunc("/addnode", s.updateAddNode).Methods("POST")
	myRouter.HandleFunc("/delnode", s.updateDelNode).Methods("POST")

	/*write new entry*/
	myRouter.HandleFunc("/data", s.doPut).Methods("POST")

	/*return single entry*/
	myRouter.HandleFunc("/data", s.doGet).Methods("GET")

	// Bucket creation
	myRouter.HandleFunc("/db/{bucketName}", s.doCreateBucket).Methods("POST")

	server := http.Server{
		Addr:    fmt.Sprintf(":%v", s.nodeId), //:{port}
		Handler: myRouter,
	}
	return &server
}

func NewNodeServer(port int64, newRing *consistenthash.Ring) *nodesServer {
	// Open database
	db, err := bolt.ConnectDB(int(port))
	if err != nil {
		log.Fatal("Error connecting to Bolt database:", err)
	}
	return &nodesServer{
		nodeId: port,
		ring: newRing,
		boltDB: db,
		repServer: newReplicaServer(db),
	}
}