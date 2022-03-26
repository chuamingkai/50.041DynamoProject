package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	consistenthash "github.com/chuamingkai/50.041DynamoProject/pkg/consistenthashing"
)

// http://localhost:8000/
// open ports 9000-9100 for client

// functions: maintain ring structure, make POST req to node http server to add/delete nodes, handles GET req by client and return the node name for client to approach (send port number first in preference list), handle POST req of new node aka when new node joins updates its ring structure

// current: https://groups.google.com/g/golang-nuts/c/80WQU8u61PI uint64 error

// let's declare a global Keys array that we can then populate in our main function to simulate a database
var ClientReq []ClientRequest
var Nodes []UpdateNode

type NodeManager struct {
	RingList consistenthash.Ring
}

var manager NodeManager

// object to send to nodes:

// ClientHTPP e.g.:{"key":"954336","type":"GET"}
// format for client to request for a node using key on GET
// RequestTypes: "GET": find node from key
type ClientRequest struct {
	Key string `json:"key"`
	// RequestType string `json:"type"`
	// NodeName string `json:"node_name,omitempty"`
}

type UpdateNode struct {
	Key    string `json:"key"`
	PortNo uint64 `json:"portno"`
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Client landing!")
	fmt.Println("Endpoint Hit: nodemanager http server")
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/allClient", returnAllClientReq)      // all client reqs
	myRouter.HandleFunc("/client", newRequest).Methods("POST") //allow client to post to manager
	myRouter.HandleFunc("/client/{key}", returnSingleKey)
	myRouter.HandleFunc("/allNodes", returnAllNodes)            // all existing nodes
	myRouter.HandleFunc("/addNode", addNodeReq).Methods("POST") // managing ADD
	myRouter.HandleFunc("/delNode", delNodeReq).Methods("POST") // managing DEL
	myRouter.HandleFunc("/delNode", deleteKey).Methods("POST")  //testing local delete
	log.Fatal(http.ListenAndServe(":8000", myRouter))
}

func addNodeReq(w http.ResponseWriter, r *http.Request) {
	// get the body of our POST request
	// unmarshal this into a new ClientRequest struct
	// append this to our ClientReq array, and check for node responsible for key
	reqBody, _ := ioutil.ReadAll(r.Body)
	var article UpdateNode

	var err error = json.Unmarshal(reqBody, &article)
	if err != nil {
		panic(err)
	}
	Nodes = append(Nodes, article)
	json.NewEncoder(w).Encode(&article)
	// add new node to the ring
	fmt.Println("Adding node", article.Key, " to existing ring.")
	manager.addNode(article.Key, article.PortNo)
}

func delNodeReq(w http.ResponseWriter, r *http.Request) {
	// get the body of our POST request
	// unmarshal this into a new ClientRequest struct
	// append this to our ClientReq array, and check for node responsible for key
	reqBody, _ := ioutil.ReadAll(r.Body)
	var article UpdateNode

	var err error = json.Unmarshal(reqBody, &article)
	if err != nil {
		panic(err)
	}

	// add new node to the ring
	fmt.Println("Removing node", article.Key, " from existing ring.")
	manager.removeNode(article.Key, article.PortNo)
}

func returnSingleKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	// Loop over all of our Articles - if the article.Id equals the key we pass in, return the article encoded as JSON
	for _, article := range ClientReq {
		if article.Key == key {
			json.NewEncoder(w).Encode(article)
		}
	}
}

func returnAllNodes(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnAllNodes")
	json.NewEncoder(w).Encode(Nodes)
}

// encode keys to arr->json str
func returnAllClientReq(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnAllClientReq")
	json.NewEncoder(w).Encode(ClientReq)
}

func newRequest(w http.ResponseWriter, r *http.Request) {
	// get the body of our POST request
	// unmarshal this into a new ClientRequest struct
	// append this to our ClientReq array, and check for node responsible for key
	reqBody, _ := ioutil.ReadAll(r.Body)
	var article ClientRequest

	var err error = json.Unmarshal(reqBody, &article)
	if err != nil {
		panic(err)
	}

	// update our global ClientReq array to include
	// our new ClientRequest

	ClientReq = append(ClientReq, article)
	json.NewEncoder(w).Encode(&article)

	// search for expected node
	fmt.Println("Searching for node for requested key", article.Key)
	//var response = manager.RingList.SearchKey(article.Key)
	var response = manager.RingList.GetPreferenceList(article.Key)

	// since we are simply emulating client, print response to terminal
	//fmt.Println("The node responsible is at Node ID:", response.NodeId,"\n================================================================")
	// get first node in preference list instead
	fmt.Println("The node responsible is at Node ID:", response[0].NodeId, "\n================================================================")
}

// unused for now
func deleteKey(w http.ResponseWriter, r *http.Request) {
	// once again, we will need to parse the path parameters
	vars := mux.Vars(r)
	// we will need to extract the `id` of the article we
	// wish to delete
	id := vars["key"]

	// we then need to loop through all our articles
	for index, key := range Nodes {
		// if our id path parameter matches one of our
		// articles
		if key.Key == id {
			// updates our Articles array to remove the
			// article
			Nodes = append(Nodes[:index], Nodes[index+1:]...)
		}
	}
}

func (m *NodeManager) removeNode(name string, id uint64) {
	fmt.Println("Removing node", name, " at port:", id)
	m.RingList.RemoveNode(name)
	fmt.Print("Deleted!")
}

func (m *NodeManager) addNode(name string, id uint64) {
	fmt.Println("Adding node", name, " at port:", id)
	m.RingList.AddNode(name, id)
	fmt.Println("added!\n================================================================")
}

func initManager() {
	manager.RingList = *consistenthash.NewRing()
}

func main() {

	nodeNames := []string{"1", "2", "3", "4"}
	nodeNumbers := []uint64{9030, 9040, 9050, 9060}
	initManager()
	fmt.Println("manager: ", manager)

	testManager(nodeNames, nodeNumbers)
	//populate with dummy data
	//ClientReq = []ClientRequest{}

	handleRequests()
}

func testManager(nodenames []string, nodeIDs []uint64) {
	for i := 0; i < len(nodenames); i++ {
		manager.addNode(nodenames[i], nodeIDs[i])
		fmt.Println("returned from adding node ", nodeIDs[i])
	}

}

/*
	// id := 55 // TODO: Dynamically assign node IDs
	testEntry := models.Object{
        Key string `json:"key"`
	    Value string `json:"value"`
	    VC     map[string]uint64 `json:",omitempty"`
		//IC: "S1234567A",
		//GeoLoc: "23:23:23:23 NW",
	}

	// Open database
	db, err := bolt.ConnectDB(id)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	defer db.DB.Close()

	// Create bucket
	err = db.CreateBucket("testBucket")
	if err != nil {
		log.Fatalf("Error creating bucket: %s", err)
	}

	// Insert test value into bucket
	err = db.Put("testBucket", testEntry)
	if err != nil {
		log.Fatalf("Error inserting into bucket: %s", err)
	}

	// Read from bucket
	value := db.Get("testBucket", testEntry.IC)
	fmt.Printf("Value at key %s: %s", testEntry.IC, value)
}
	// getResult := db.Get("testBucket", testEntry.Key)
	// fmt.Println(getResult)

*/
