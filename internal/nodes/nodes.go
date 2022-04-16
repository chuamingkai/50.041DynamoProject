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

	config "github.com/chuamingkai/50.041DynamoProject/config"
	"github.com/chuamingkai/50.041DynamoProject/internal/models"
	"github.com/chuamingkai/50.041DynamoProject/pkg/bolt"
	consistenthash "github.com/chuamingkai/50.041DynamoProject/pkg/consistenthashing"
	pb "github.com/chuamingkai/50.041DynamoProject/pkg/internalcomm"

	"github.com/gorilla/mux"
)

type nodesServer struct {
	pb.UnimplementedReplicationServer
	boltDB       *bolt.DB
	ring         *consistenthash.Ring
	nodeId       int64
	internalAddr int64
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
	Error         string `json:"error"`
}

type GetRequestBody struct {
	BucketName string `json:"bucketName"`
	Key        string `json:"key"`
}

type GetResponseBody struct {
	Context   models.GetContext `json:"context"`
	Value     string            `json:"value,omitempty"`
	Conflicts []models.Object   `json:"conflicts,omitempty"`
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

	// Check if bucket name is the protected hints bucket
	if reqBody.BucketName == config.HINT_BUCKETNAME {
		http.Error(w, fmt.Sprintf("Read/Write to '%s' bucket not allowed", config.HINT_BUCKETNAME), http.StatusForbidden)
		return
	}

	// Check if node handles key
	if !s.ring.IsNodeResponsibleForKey(reqBody.Object.Key, uint64(s.internalAddr)) {
		http.Error(w, "Node is not responsible for key!", http.StatusBadRequest)
		return
	}

	log.Printf("Received PUT request from clients for key %s, value %v\n", reqBody.Object.Key, reqBody.Object.Value)

	dataBytes, err := json.Marshal(reqBody.Object)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	putRepSuccessStatus, err := s.serverPutReplicas(reqBody.BucketName, reqBody.Object.Key, dataBytes)
	switch putRepSuccessStatus {
	case FULL_SUCCESS:
		w.WriteHeader(http.StatusCreated)
	case PARTIAL_SUCCESS:
		w.WriteHeader(http.StatusCreated)
		putResponseBody := PutResponseBody{
			SuccessStatus: "PARTIAL SUCCESS",
			Error:         err.Error(),
		}
		putResponseBodyBytes, _ := json.Marshal(putResponseBody)
		w.Write(putResponseBodyBytes)
	case UNSUCCESSFUL:
		w.WriteHeader(http.StatusInternalServerError)
		putResponseBody := PutResponseBody{
			SuccessStatus: "UNSUCCESSFUL",
			Error:         err.Error(),
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

	// Check if bucket name is the protected hints bucket
	if reqBody.BucketName == config.HINT_BUCKETNAME {
		http.Error(w, fmt.Sprintf("Read/Write to '%s' bucket not allowed", config.HINT_BUCKETNAME), http.StatusForbidden)
		return
	}

	// Check if node handles key
	if !s.ring.IsNodeResponsibleForKey(reqBody.Key, uint64(s.internalAddr)) {
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
			HasConflicts:     getRepResult.HasVersionConflict,
			WriteCoordinator: getRepResult.WriteCoordinator,
		}
	} else {
		obj := getRepResult.Replicas[0]

		getResponseBody.Value = obj.Value
		context = models.GetContext{
			HasConflicts:     getRepResult.HasVersionConflict,
			WriteCoordinator: getRepResult.WriteCoordinator,
			VectorClock:      obj.VC,
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
	// protected hint bucket name
	if newBucketName == config.HINT_BUCKETNAME {
		http.Error(w, fmt.Sprintf("'%s' cannot be the name of a bucket", config.HINT_BUCKETNAME), http.StatusBadRequest)
		return
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

	// Change node name to gRPC address of node
	grpcAddress := newnode.NodeName - 3000

	for _, n := range s.ring.NodeMap {
		if n.NodeId == grpcAddress {
			exist = true
		}
		break
	}

	if !exist {
		//s.ring.AddNode(nodenameString, grpcAddress)
		s.serverReallocKeys(nodenameString, config.MAIN_BUCKETNAME, grpcAddress)

		successMessage := fmt.Sprintf("Node %v added to ring.", newnode.NodeName)
		log.Println(successMessage)
		log.Println(s.ring.Nodes.TraverseAndPrint())
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(successMessage)
	} else {
		http.Error(w, "Node already exists!", http.StatusBadRequest)
	}
}

func (s *nodesServer) confirmDelNode(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)

	var node UpdateNodeRequestBody
	if err := json.Unmarshal(reqBody, &node); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.delnodeReallocKeys(strconv.FormatUint(node.NodeName, 10))
}

// TODO: What happens when node receives delete request for itself?
func (s *nodesServer) updateDelNode(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)

	var newnode UpdateNodeRequestBody
	if err := json.Unmarshal(reqBody, &newnode); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	exist := false
	internalNodename := newnode.NodeName - 3000
	externalNodenameString := strconv.FormatUint(newnode.NodeName, 10)

	for _, n := range s.ring.NodeMap {
		if n.NodeId == internalNodename {
			exist = true
		}
	}
	if exist {
		successMessage := fmt.Sprintf("Node %v removed from ring.", newnode.NodeName)
		s.ring.RemoveNode(externalNodenameString)
		log.Println(successMessage)
		log.Println(s.ring.Nodes.TraverseAndPrint())
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(successMessage)
	} else {
		http.Error(w, "Node doesn't exist!", http.StatusBadRequest)
	}
}

// Get all objects inside a bucket
func (s *nodesServer) doGetAllBucketObjects(w http.ResponseWriter, r *http.Request) {
	pathVars := mux.Vars(r)
	var ok bool
	var bucketName string
	if bucketName, ok = pathVars["bucketName"]; !ok {
		http.Error(w, "Missing bucket name", http.StatusBadRequest)
		return
	}

	bucketObjectsBytes, err := s.boltDB.GetAllObjects(bucketName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var bucketObjects []models.Object
	for _, objBytes := range bucketObjectsBytes {
		var obj models.Object
		if err := json.Unmarshal(objBytes, &obj); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		bucketObjects = append(bucketObjects, obj)
	}

	json.NewEncoder(w).Encode(bucketObjects)
}

// Get all buckets inside the database
func (s *nodesServer) getAllBuckets(w http.ResponseWriter, r *http.Request) {
	bucketNames, err := s.boltDB.GetAllBuckets()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		json.NewEncoder(w).Encode(bucketNames)
	}
}

// ringBackupService is a function that runs periodically backups the ring
// by calling BackupRing to a file, specified by RING_BACKUP_FILEFORMAT
// Interval to backup defined by RING_BACKUP_INTERVAL in config file
// Filename of backup defined by RING_BACKUP_FILEFORMAT in config file
func (s *nodesServer) ringBackupService() {
	backupTicker := time.NewTicker(config.RING_BACKUP_INTERVAL)
	for {
		<-backupTicker.C
		err := s.ring.BackupRing(fmt.Sprintf(config.RING_BACKUP_FILEFORMAT, s.nodeId))
		if err != nil {
			log.Printf("RingBackupService error: %s\n", err)
		}
	}
}

// hintHandlerService periodically checks the 'hints' bucket of the database
// and tries to pass the hinted value to the appropriate node.
// Interval check defined by HINT_CHECK_INTERVAL in config file
func (s *nodesServer) hintHandlerService() {
	ticker := time.NewTicker(config.HINT_CHECK_INTERVAL)
	for {
		<-ticker.C
		delete := make([]string, 0)
		err := s.boltDB.Iterate(config.HINT_BUCKETNAME, func(k, v []byte) error {
			destinationNodename := string(k[:])
			targetPort := s.ring.NodeMap[destinationNodename].NodeId
			if s.sendHeartbeat(targetPort) {
				log.Printf("HintHandlerService: node %s responded to heartbeat\n", destinationNodename)
				var hintedDatas []models.HintedObject
				if err := json.Unmarshal(v, &hintedDatas); err != nil {
					return err
				}
				if s.clientPutMultiple(targetPort, hintedDatas) {
					delete = append(delete, destinationNodename)
				}
			} else {
				log.Printf("HintHandlerService: node %s failed to respond to heartbeat\n", destinationNodename)
			}
			return nil
		})

		if err != nil {
			log.Printf("HintHandlerService error: %s\n", err)
		}

		for _, v := range delete {
			err = s.boltDB.DeleteKey(config.HINT_BUCKETNAME, v)
			log.Printf("HintHandlerService: Hint successfully handed off to node %s\n", v)
			if err != nil {
				log.Printf("HintHandlerService error: %s\n", err)
			}
		}
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
	myRouter.HandleFunc("/confirmdel", s.confirmDelNode).Methods("POST")

	// Write new entry
	myRouter.HandleFunc("/data", s.doPut).Methods("POST")

	// Return single entry
	myRouter.HandleFunc("/data", s.doGet).Methods("GET")

	// Get all bucket names
	myRouter.HandleFunc("/db", s.getAllBuckets).Methods("GET")

	// Read from bucket
	myRouter.HandleFunc("/db/{bucketName}", s.doGetAllBucketObjects).Methods("GET")

	// Bucket creation
	myRouter.HandleFunc("/db/{bucketName}", s.doCreateBucket).Methods("POST")

	server := http.Server{
		Addr:    fmt.Sprintf(":%v", s.nodeId), //:{port}
		Handler: myRouter,
	}

	go s.ringBackupService()
	go s.hintHandlerService()
	return &server, nil
}

func NewNodeServer(port, internalAddr int64, newRing *consistenthash.Ring) *nodesServer {
	// Open database
	db, err := bolt.ConnectDB(int(port))
	if err != nil {
		log.Fatal("Error connecting to Bolt database:", err)
	}

	if !db.BucketExists(config.HINT_BUCKETNAME) {
		err = db.CreateBucket(config.HINT_BUCKETNAME)
		if err != nil {
			log.Fatal("Error creating hint bucket:", err)
		}
	}
	/*if !db.BucketExists(config.MAIN_BUCKETNAME) {
		err = db.CreateBucket(config.MAIN_BUCKETNAME)
		if err != nil {
			log.Fatal("Error creating main bucket:", err)
		}
	}*/
	return &nodesServer{
		nodeId:       port,
		internalAddr: internalAddr,
		ring:         newRing,
		boltDB:       db,
	}
}
