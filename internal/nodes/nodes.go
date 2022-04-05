package nodes

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/chuamingkai/50.041DynamoProject/internal/models"
	"github.com/chuamingkai/50.041DynamoProject/pkg/bolt"
	consistenthash "github.com/chuamingkai/50.041DynamoProject/pkg/consistenthashing"
	pb "github.com/chuamingkai/50.041DynamoProject/pkg/internalcomm"

	"github.com/gorilla/mux"
)

type nodesServer struct {
	pb.UnimplementedReplicationServer
	boltDB 			*bolt.DB
	ring   			*consistenthash.Ring
	nodeId 			int64
	internalAddr 	int64
}
type UpdateNodeRequestBody struct {
	NodeName uint64 `json:"nodename"`
}

type PutRequestBody struct {
	BucketName string         `json:"bucketName"`
	Object     *models.Object `json:"object"`
}

// Only for partially successful/uncsuccessful PUT operations
type PutResponseBody struct {
	SuccessStatus string `json:"successStatus"`
	Error string `json:"error"`
}

type GetRequestBody struct {
	BucketName string `json:"bucketName"`
	Key        string `json:"key"`
}

type GetResponseBody struct {
	Context 		models.GetContext 	`json:"context"`
	Value 			string 				`json:"value,omitempty"`
	Conflicts 		[]models.Object 	`json:"conflicts,omitempty"`
	SuccessStatus 	SuccessStatus 		`json:"successStatus"`
}

func (s *nodesServer) doPut(w http.ResponseWriter, r *http.Request) {
	reqBodyBytes, _ := ioutil.ReadAll(r.Body)
	var reqBody PutRequestBody
	json.Unmarshal(reqBodyBytes, &reqBody)

	// Check if bucket name is provided
	if reqBody.BucketName == "" {
		http.Error(w, "Missing bucket name", http.StatusBadRequest)
		return
	}

	// Check if node handles key
	if !s.ring.IsNodeResponsibleForKey(reqBody.Object.Key, uint64(s.nodeId)) {
		http.Error(w, "Node is not responsible for key!", http.StatusBadRequest)
		return
	}

	log.Printf("Received PUT request from clients for key %s, value %v\n", reqBody.Object.Key, reqBody.Object.Value)

	dataBytes, err := json.Marshal(reqBody.Object)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	putRepSuccessStatus, err := s.serverPutReplicas(reqBody.BucketName, reqBody.Object.Key, dataBytes)
	switch(putRepSuccessStatus) {
	case FULL_SUCCESS:
		w.WriteHeader(http.StatusCreated)
	case PARTIAL_SUCCESS:
		w.WriteHeader(http.StatusCreated)
		putResponseBody := PutResponseBody{
			SuccessStatus: "PARTIAL SUCCESS",
			Error: err.Error(),
		}
		putResponseBodyBytes, _ := json.Marshal(putResponseBody)
		w.Write(putResponseBodyBytes)
	case UNSUCCESSFUL:
		w.WriteHeader(http.StatusInternalServerError)
		putResponseBody := PutResponseBody{
			SuccessStatus: "UNSUCCESSFUL",
			Error: err.Error(),
		}
		putResponseBodyBytes, _ := json.Marshal(putResponseBody)
		w.Write(putResponseBodyBytes)
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
	if !s.ring.IsNodeResponsibleForKey(reqBody.Key, uint64(s.nodeId)) {
		http.Error(w, "Node is not responsible for key!", http.StatusBadRequest)
		return
	}

	log.Printf("Received GET request from clients for key %s\n", reqBody.Key)

	getRepResult, err := s.serverGetReplicas(reqBody.BucketName, reqBody.Key)

	if err != nil {
		log.Printf("Error getting replicas: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} 
	
	if len(getRepResult.Replicas) == 0 {
		http.Error(w, fmt.Sprintf("key {%v} not found in bucket {%v}", reqBody.Key, reqBody.BucketName), http.StatusNotFound)
		return
	}
	
	var getResponseBody GetResponseBody
	var context models.GetContext
	if getRepResult.HasVersionConflict {
		getResponseBody.Conflicts = getRepResult.Replicas
		context = models.GetContext{
			HasConflicts: getRepResult.HasVersionConflict,
			WriteCoordinator: getRepResult.WriteCoordinator,
		}
	} else {
		obj := getRepResult.Replicas[0]

		getResponseBody.Value = obj.Value
		context = models.GetContext{
			HasConflicts: getRepResult.HasVersionConflict,
			WriteCoordinator: getRepResult.WriteCoordinator,
			VectorClock: obj.VC,
		}
	}

	getResponseBody.Context = context
	json.NewEncoder(w).Encode(getResponseBody)

}

func (s *nodesServer) doCreateBucket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var newBucketName string
	var ok bool
	if newBucketName, ok = vars["bucketName"]; !ok {
		http.Error(w, "Missing bucket name", http.StatusBadRequest)
	}

	err := s.boltDB.CreateBucket(newBucketName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
}

func (s *nodesServer) updateAddNode(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)

	var newnode UpdateNodeRequestBody
	json.Unmarshal(reqBody, &newnode)
	exist := false
	nodenameString := strconv.FormatUint(newnode.NodeName, 10)

	// Change node id to gRPC address of node
	// newnode.NodeName -= 3000
	grpcAddress := newnode.NodeName - 3000

	for _, n := range s.ring.NodeMap {
		if n.NodeId == uint64(newnode.NodeName) {
			exist = true
		}
		break
	}

	if !exist {
		s.ring.AddNode(nodenameString, grpcAddress)
		log.Printf("Node %v added to ring.\n", newnode.NodeName)
		// fmt.Println(s.ring.Nodes.TraverseAndPrint())
		json.NewEncoder(w).Encode(s.ring.Nodes.TraverseAndPrint())
	} else {
		http.Error(w, "Node already exists!", http.StatusBadRequest)
	}
}

func (s *nodesServer) updateDelNode(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)

	var newnode UpdateNodeRequestBody
	json.Unmarshal(reqBody, &newnode)
	exist := false
	nodenameString := strconv.FormatUint(newnode.NodeName, 10)

	for _, n := range s.ring.NodeMap {
		if n.NodeId == newnode.NodeName {
			exist = true
		}
	}
	if exist {
		s.ring.RemoveNode(nodenameString)
		fmt.Println(s.ring.Nodes.TraverseAndPrint())
		json.NewEncoder(w).Encode(newnode)
	} else {
		http.Error(w, "Node doesn't exists!", http.StatusServiceUnavailable)
	}
}

func (s *nodesServer) RunNodeServer() (*http.Server, error) {
	if s.ring == nil {
		return nil, errors.New("nil ring pointer")
	}

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
	return &server, nil
}

func (s *nodesServer) ringBackupService() {
	backupTicker := time.NewTicker(1 * time.Minute) // randomly chosen time
	for {
		<-backupTicker.C
		s.ring.BackupRing(fmt.Sprintf("store/ringbackups/node%dring.bak", s.nodeId))
	}
}

func NewNodeServer(port, internalAddr int64, newRing *consistenthash.Ring) *nodesServer {
	// Open database
	db, err := bolt.ConnectDB(int(port))
	if err != nil {
		log.Fatal("Error connecting to Bolt database:", err)
	}
	return &nodesServer{
		nodeId: port,
		internalAddr: internalAddr,
		ring:   newRing,
		boltDB: db,
	}
}
