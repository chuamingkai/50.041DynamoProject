package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/chuamingkai/50.041DynamoProject/config"
	consistenthash "github.com/chuamingkai/50.041DynamoProject/pkg/consistenthashing"
)

// http://localhost:8000/
// open ports 9000-9100 for client

// functions: maintain ring structure, make POST req to node http server to add/delete nodes, handles GET req by client and return the node name for client to approach (send port number first in preference list), handle POST req of new node aka when new node joins updates its ring structure

// declare a global Keys array that we can then populate in our main function to simulate a database for client request
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
	//Key    string `json:"key"`
	//PortNo uint64 `json:"portno"`
	NodeName uint64 `json:"nodename"`
}

// port 8000 home page display
func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Node manager landing @ 8000!\n")
	fmt.Fprintf(w, "For clients, make a GET request to '/getKey' with a key. The manager will return the node you will need to communicate with.\n")
	fmt.Fprintf(w, "To add a node to the ring, make a POST request to '/addNode' with a key and port number.\n")
	fmt.Fprintf(w, "To remove a node from the ring, make a POST request to '/removeNode' with a key and port number.\n")
	fmt.Println("Endpoint Hit: nodemanager http server")
}

func (m *NodeManager) handleRequests() {

	fmt.Println("Node manager started at 8000!")
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/allClient", returnAllClientReq)
	myRouter.HandleFunc("/getKey", m.getKey).Methods("GET") // allow client to make GET req for key and get response
	myRouter.HandleFunc("/client/{key}", returnSingleKey)
	myRouter.HandleFunc("/addNode", m.addNodeReq).Methods("POST") // managing ADD
	myRouter.HandleFunc("/delNode", m.delNodeReq).Methods("POST") // managing DEL

	log.Fatal(http.ListenAndServe(":8000", myRouter))
}

// respond to the requesting add node with the current ring
func (m *NodeManager) addNodeReq(w http.ResponseWriter, r *http.Request) {

	reqBody, _ := ioutil.ReadAll(r.Body)
	var newnode UpdateNode
	var listofexistingnode []UpdateNode

	var err error = json.Unmarshal(reqBody, &newnode)
	if err != nil {
		log.Fatal()
	}
	//Nodes = append(Nodes, article)
	//json.NewEncoder(w).Encode(&newnode)

	exist := false
	for _, n := range m.RingList.NodeMap {
		fmt.Println(n.NodeId, newnode.NodeName)
		if n.NodeId == newnode.NodeName {
			fmt.Println("node found:", newnode.NodeName)
			exist = true
			break
		}
	}

	// Make request to nodes to addnode
	if !exist {
		if len(m.RingList.NodeMap) >= 1 {
			for _, node := range m.RingList.NodeMap {
				fmt.Println("Requesting to add node to", newnode.NodeName)
				postBody, _ := json.Marshal(newnode)
				responseBody := bytes.NewBuffer(postBody)
				resp, err := http.Post(fmt.Sprintf("http://localhost:%v/addnode", node.NodeId), "application/json", responseBody)
				if err != nil {
					// handle error
					panic(err)
				}
				// close response body when finished with it
				defer resp.Body.Close()
				listofexistingnode = append(listofexistingnode, UpdateNode{NodeName: node.NodeId})
				//body, err := io.ReadAll(resp.Body)
				//if err != nil {
				//	panic(err)
				//}
				// this body is unreadable/may need another conversion first. Using simple string(body) causes a parsing error so this is roundabout
				//strBody := strconv.FormatUint(binary.LittleEndian.Uint64(body), 16)
				//log.Println(strBody)
				//fmt.Println("Returned from adding.")
				//fmt.Println("Response body addNode",string(body))
			}
		}
		// add new node to the ring
		fmt.Println("Adding node", newnode.NodeName, " to existing ring.")
		stringName := strconv.FormatUint(newnode.NodeName, 10)
		m.addNode(stringName, newnode.NodeName)
		//json.NewEncoder(w).Encode(m.RingList.Nodes.TraverseAndPrint())
		fmt.Println(listofexistingnode)
		json.NewEncoder(w).Encode(listofexistingnode)
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
		log.Fatal()
	}

	exist := false
	for _, n := range m.RingList.NodeMap {
		if n.NodeId == delNode.NodeName {
			exist = true
			break
		}
	}
	if exist {
		fmt.Println("Removing node", delNode.NodeName, " from manager's ring.")
		stringName := strconv.FormatUint(delNode.NodeName, 10)
		m.RingList.RemoveNode(stringName)

		/*send http post request to all nodes it knows*/
		if m.RingList.Nodes.Length >= 1 {

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
				//body, err := ioutil.ReadAll(resp.Body)
				//if err != nil {
				//	panic(err)
				//}
				// this body is unreadable/may need another conversion first. Using simple string(body) causes a parsing error so this is roundabout
				//strBody := strconv.FormatUint(binary.LittleEndian.Uint64(body), 16)
				//log.Println(strBody)
				//fmt.Println("Response body addNode",string(body))
				//fmt.Println("Returned from deletion")

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

// displays all client reqs to date
func returnAllClientReq(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnAllClientReq")
	json.NewEncoder(w).Encode(ClientReq)
}

func (m *NodeManager) getKey(w http.ResponseWriter, r *http.Request){
	// allow client to make GET request
	reqBodyBytes, _ := ioutil.ReadAll(r.Body)
	var reqBody ClientRequest
	json.Unmarshal(reqBodyBytes, &reqBody)

	// Check if key is provided
	if reqBody.Key == "" {
		http.Error(w, "Missing key", http.StatusBadRequest)
		return
	}

	// update our global ClientReq array to include
	// our new ClientRequest
	ClientReq = append(ClientReq, reqBody)
	json.NewEncoder(w).Encode(&reqBody)

	// search for expected node on pref list
	fmt.Println("Client request received. Searching for node for requested key", reqBody.Key)
	var response = m.RingList.GetPreferenceList(reqBody.Key)
	
	var respBody UpdateNode
	respBody.NodeName = response[0].NodeId

	// since we are simply emulating client, print response to terminal
	fmt.Println("The node responsible is at Node ID:", response[0].NodeId, "\n================================================================")

	json.NewEncoder(w).Encode(respBody)

}


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
	if err := config.LoadEnvFile(); err != nil {
		log.Fatalf("Error loading .env file: %v\n", err.Error())
	}
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
