package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"encoding/binary"

	"github.com/gorilla/mux"

	consistenthash "github.com/chuamingkai/50.041DynamoProject/pkg/consistenthashing"
)

// http://localhost:8000/
// open ports 9000-9100 for client

// functions: maintain ring structure, make POST req to node http server to add/delete nodes, handles GET req by client and return the node name for client to approach (send port number first in preference list), handle POST req of new node aka when new node joins updates its ring structure

// let's declare a global Keys array that we can then populate in our main function to simulate a database for client request
var ClientReq []ClientRequest
var Nodes []UpdateNode //not displayed

type NodeManager struct {
	nodeID   int64
	RingList *consistenthash.Ring
}

// ClientHTPP e.g.:{"key":"A1234"}
// format for client to request for a node using key on GET
type ClientRequest struct {
	Key string `json:"key"`
}

// object to send to nodes:
type UpdateNode struct {
	Key    string `json:"key"`
	PortNo uint64 `json:"portno"`
}

// port 8000 home page display
func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Node manager landing @ 8000!\n")
	fmt.Fprintf(w, "For clients, make a POST request to '/client' with a key. The manager will return the node you will need to communicate with.\n")
	fmt.Fprintf(w, "To add a node to the ring, make a POST request to '/addNode' with a key and port number.\n")
	fmt.Fprintf(w, "To remove a node from the ring, make a POST request to '/removeNode' with a key and port number.\n")
	fmt.Println("Endpoint Hit: nodemanager http server")
}

func (m *NodeManager) handleRequests()  {

	fmt.Println("Node manager started at 8000!")
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/allClient", returnAllClientReq)
	myRouter.HandleFunc("/client", m.newRequest).Methods("POST") //allow client to post to manager to get node portno
	myRouter.HandleFunc("/client/{key}", returnSingleKey)
	//myRouter.HandleFunc("/allNodes", returnAllNodes)            // all existing nodes -- doesnt work yet
	myRouter.HandleFunc("/addNode", m.addNodeReq).Methods("POST") // managing ADD
	myRouter.HandleFunc("/delNode", m.delNodeReq).Methods("POST") // managing DEL
	//myRouter.HandleFunc("/delNode", deleteKey).Methods("POST")  // testing local array delete

	log.Fatal(http.ListenAndServe(":8000", myRouter))
}

// respond to the requesting add node with the current ring
func (m *NodeManager) addNodeReq(w http.ResponseWriter, r *http.Request) {

	reqBody, _ := ioutil.ReadAll(r.Body)
	var article UpdateNode

	var err error = json.Unmarshal(reqBody, &article)
	if err != nil {
		panic(err)
	}
	//Nodes = append(Nodes, article)
	json.NewEncoder(w).Encode(&article)
	// add new node to the ring
	fmt.Println("Adding node", article.PortNo, " to existing ring.")
	m.addNode(article.Key, article.PortNo)

	exist := false
	var newnode UpdateNode
	for _, n := range m.RingList.NodeMap {
		fmt.Println(n.NodeId, newnode.PortNo)
		if n.NodeId == newnode.PortNo {
			exist = true
		}
	}
	// Make request to nodes to addnode
	if !exist {
		if len(m.RingList.NodeMap) > 1 {
			for _, node := range m.RingList.NodeMap {
				fmt.Println("Requesting to add node to", article.PortNo)
				postBody, _ := json.Marshal(newnode)
				responseBody := bytes.NewBuffer(postBody)
				resp, err := http.Post(fmt.Sprintf("http://localhost:%v/addnode", node.NodeId), "application/json", responseBody)
				if err != nil {
					// handle error
					continue
				}
				// close response body when finished with it
				defer resp.Body.Close()
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					continue
				}
				// this body is unreadable/may need another conversion first. Using simple string(body) causes a parsing error so this is roundabout
				strBody := strconv.FormatUint(binary.LittleEndian.Uint64(body),16)
				log.Println(strBody)
				//fmt.Println("Response body addNode",string(body))
			}
		}
		json.NewEncoder(w).Encode(m.RingList.Nodes.TraverseAndPrint())
	} else {
		fmt.Println("Failed to add a node: already exists.")
		http.Error(w, "Node already exists!", http.StatusBadRequest)
	}
}

func (m *NodeManager) delNodeReq(w http.ResponseWriter, r *http.Request) {
	// get the body of our POST request
	// unmarshal this into a new ClientRequest struct
	// append this to our ClientReq array, and check for node responsible for key
	reqBody, _ := ioutil.ReadAll(r.Body)
	var delNode UpdateNode

	var err error = json.Unmarshal(reqBody, &delNode)
	if err != nil {
		panic(err)
	}

	// del node from the ring
	fmt.Println("Removing node", delNode.Key, " from ring.")
	//manager.removeNode(delNode.Key, delNode.PortNo)

	exist := false
	for _, n := range m.RingList.NodeMap {
		if n.NodeId == delNode.PortNo {
			exist = true
		}
	}
	if exist {
		m.RingList.RemoveNode(delNode.Key)

		/*send http post request to all nodes it knows*/
		if m.RingList.Nodes.Length > 1 {

			for _, node := range m.RingList.NodeMap {

				postBody, _ := json.Marshal(delNode)
				responseBody := bytes.NewBuffer(postBody)
				// make Req to nodes to delete
				resp, err := http.Post(fmt.Sprintf("http://localhost:%v/delnode", node.NodeId), "application/json", responseBody)
				//Handle Error
				if err != nil {
					log.Fatalf("An Error Occured %v", err)
				}
				// close response body when finished with it
				defer resp.Body.Close()
				//Read the response body
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					continue
				}
				// this body is unreadable/may need another conversion first. Using simple string(body) causes a parsing error so this is roundabout
				strBody := strconv.FormatUint(binary.LittleEndian.Uint64(body),16)
				log.Println(strBody)
				//fmt.Println("Response body addNode",string(body))

			}
		}

		fmt.Println(m.RingList.Nodes.TraverseAndPrint())
	} else {
		fmt.Println("Failed to delete node: node doesn't exist.")
		http.Error(w, "Node doesn't exist!", http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(delNode)
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

// unused - to display all currently listed nodes on manager
func returnAllNodes(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnAllNodes")
	json.NewEncoder(w).Encode(Nodes)
}

// displays all client reqs to date
func returnAllClientReq(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnAllClientReq")
	json.NewEncoder(w).Encode(ClientReq)
}

func (m *NodeManager) newRequest(w http.ResponseWriter, r *http.Request) {
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

	// search for expected node on pref list
	fmt.Println("Searching for node for requested key", article.Key)
	var response = m.RingList.GetPreferenceList(article.Key)

	// since we are simply emulating client, print response to terminal
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
/*
func (m *NodeManager) removeNode(name string, id uint64) {
	fmt.Println("Removing node", name, " at port:", id)
	//m.RingList.RemoveNode(name)
	fmt.Print("Deleted! Current nodes: ")
	//fmt.Println(m.RingList.Nodes.TraverseAndPrint())
}*/

func (m *NodeManager) addNode(name string, id uint64) {
	fmt.Println("Adding node", name, " at port:", id)
	m.RingList.AddNode(name, id)
	fmt.Println("Added! Current nodes: ")
	fmt.Println(m.RingList.Nodes.TraverseAndPrint())
	fmt.Println("\n================================================================")
}

func initManager(port int64, newRing *consistenthash.Ring) *NodeManager {
	return &NodeManager{
		nodeID:   port,
		RingList: newRing,
	}
}

func main() {

	//nodeNames := []string{"1", "2", "3", "4"}
	//nodeNumbers := []uint64{9030, 9040, 9050, 9060}

	managerRing := consistenthash.NewRing()
	m := initManager(8000, managerRing)

	//testManager(nodeNames, nodeNumbers)
	//populate with dummy data
	//ClientReq = []ClientRequest{}
	m.handleRequests()

}
/*
func (m *NodeManager) testManager(nodenames []string, nodeIDs []uint64) {
	for i := 0; i < len(nodenames); i++ {
		m.addNode(nodenames[i], nodeIDs[i])
		fmt.Println("Returned from adding node ", nodeIDs[i], " to local list.")
	}
}
*/