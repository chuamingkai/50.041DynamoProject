package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/chuamingkai/50.041DynamoProject/internal/models"
	"github.com/chuamingkai/50.041DynamoProject/pkg/bolt"
	consistenthash "github.com/chuamingkai/50.041DynamoProject/pkg/consistenthashing"
	pb "github.com/chuamingkai/50.041DynamoProject/pkg/internalcomm"
	"github.com/chuamingkai/50.041DynamoProject/pkg/vectorclock"

	"github.com/gorilla/mux"
)

// TODO: Name this smth else
type nodesServer struct {
	pb.UnimplementedReplicationServer
	boltDB *bolt.DB
	ring   *consistenthash.Ring
	nodeId int64
	// repServer *replicaServer
}
type UpdateNode struct {
	NodeName string `json:"nodename"`
}

type PutRequestBody struct {
	BucketName string        `json:"bucketName"`
	Object     models.Object `json:"object"`
}

type GetRequestBody struct {
	BucketName string `json:"bucketName"`
	Key        string `json:"key"`
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

	/* TODO: To add ignore if new updated vector clock is outdated?*/

	serverDeadline := time.Now().Add(10 * time.Second)
	ctxRep, cancel := context.WithDeadline(context.Background(), serverDeadline)
	defer cancel()

	dataBytes, err := json.Marshal(reqBody.Object)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	putRepReq := &pb.PutRepRequest{
		Key:        reqBody.Object.Key,
		BucketName: reqBody.BucketName,
		Data:       dataBytes,
	}

	putRepResult := s.serverPutReplicas(ctxRep, putRepReq)

	if putRepResult.Success {
		json.NewEncoder(w).Encode(reqBody.Object)
	} else {
		http.Error(w, putRepResult.Err.Error(), http.StatusInternalServerError)
	}
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

	// Check if node handles key
	nodename := strings.Trim(r.Host, "localhost:")
	number, _ := strconv.ParseUint(nodename, 10, 64)

	if !s.ring.IsNodeResponsibleForKey(reqBody.Key, number) {
		http.Error(w, "Node is not responsible for key!", http.StatusBadRequest)
		return
	}

	// Get replicas from servers in preference list
	serverDeadline := time.Now().Add(10 * time.Second)
	ctxRep, cancel := context.WithDeadline(context.Background(), serverDeadline)
	defer cancel()

	getRepReq := pb.GetRepRequest{
		BucketName: reqBody.BucketName,
		Key:        reqBody.Key,
	}

	getRepResult, err := s.serverGetReplicas(ctxRep, &getRepReq)

	if err != nil {
		log.Printf("Error getting replicas: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		json.NewEncoder(w).Encode(getRepResult)
	}

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
			log.Printf("Node %s added to ring.\n", newnode.NodeName)
			// fmt.Println(s.ring.Nodes.TraverseAndPrint())
			json.NewEncoder(w).Encode(s.ring.Nodes.TraverseAndPrint())
		} else {
			http.Error(w, "Node already exists!", http.StatusBadRequest)
		}
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

func (s *nodesServer) CreateServer() *http.Server {
	log.Println("Setting up node at port", s.nodeId)

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

	go s.ringBackupService()
	return &server
}

func (s *nodesServer) ringBackupService() {
	backupTicker := time.NewTicker(1 * time.Minute) // randomly chosen time
	for {
		<-backupTicker.C
		s.ring.BackupRing(fmt.Sprintf("store/node%dring.bak", s.nodeId))
	}
}

func NewNodeServer(port int64, newRing *consistenthash.Ring) *nodesServer {
	// Open database
	db, err := bolt.ConnectDB(int(port))
	if err != nil {
		log.Fatal("Error connecting to Bolt database:", err)
	}
	return &nodesServer{
		nodeId: port,
		ring:   newRing,
		boltDB: db,
	}
}
