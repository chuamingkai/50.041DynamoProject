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

type UpdateNode struct {
	NodeName string `json:"nodename"`
}

var a *consistenthash.Ring

func createSingleEntry(w http.ResponseWriter, r *http.Request) {

	/*To add check for if node handles key*/
	nodename := strings.Trim(r.Host, "localhost:")

	reqBody, _ := ioutil.ReadAll(r.Body)
	var data map[string]string
	json.Unmarshal(reqBody, &data)
	fmt.Println(data)

	var newEntry models.Object
	for k, v := range data {
		fmt.Printf("Key: %s, Value: %s\n", k, v)
		newEntry.Key = k
		newEntry.Value = v
	}

	number, err := strconv.ParseUint(nodename, 10, 64)
	if !a.IsNodeResponsibleForKey(newEntry.Key, number) {
		http.Error(w, "Node is not responsible for key!", 400)
		return
	}

	if newEntry.VC != nil {
		newEntry.VC = vectorclock.UpdateRecv(nodename, newEntry.VC)
	} else {
		newEntry.VC = vectorclock.UpdateRecv(nodename, map[string]uint64{nodename: 0})
	}

	/*To add ignore if new updated vector clock is outdated?*/

	// Insert test value into bucket
	err = db.Put("testBucket", newEntry)
	if err != nil {
		log.Fatalf("Error inserting into bucket: %s", err)
	}

	/*To add gRPC contact*/
	/*record first response-> writeCoordinator*/

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

	/*To add gRPC contact*/
	/*collate conflicting list*/

	/*Edit to return a list of objects instead*/
	if value.Key != "" {
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