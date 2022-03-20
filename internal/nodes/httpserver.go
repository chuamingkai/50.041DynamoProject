package nodes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/chuamingkai/50.041DynamoProject/internal/bolt"
	"github.com/chuamingkai/50.041DynamoProject/internal/models"
	"github.com/chuamingkai/50.041DynamoProject/pkg/vectorclock"

	"github.com/gorilla/mux"
)

var db bolt.DB

func createSingleEntry(w http.ResponseWriter, r *http.Request) {

	/*To add check for if node handles key*/
	nodename := strings.Trim(r.Host, "localhost:")
	// id, _ := strconv.Atoi(nodename)

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

	if newEntry.VC != nil {
		newEntry.VC = vectorclock.UpdateRecv(nodename, newEntry.VC)
	} else {
		newEntry.VC = vectorclock.UpdateRecv(nodename, map[string]uint64{nodename: 0})
	}

	/*To add ignore if new updated vector clock is outdated?*/

	// Insert test value into bucket
	err := db.Put("testBucket", newEntry)
	if err != nil {
		log.Fatalf("Error inserting into bucket: %s", err)
	}

	/*To add gRPC contact*/
	/*record first response-> writeCoordinator*/

	/*Edit to return a list of objects instead*/
	json.NewEncoder(w).Encode(newEntry)
}

func returnSingleEntry(w http.ResponseWriter, r *http.Request) {
	/*To add check for if node handles key*/
	// nodename := strings.Trim(r.Host, "localhost:")
	// id, _ := strconv.Atoi(nodename)

	vars := mux.Vars(r)
	key := vars["ic"]
	fmt.Println("Finding value at key", key)

	// Read from bucket
	value, err := db.Get("testBucket", key)
	if err != nil {
		log.Fatal("Error getting from bucket:", err)
	}

	/*To add gRPC contact*/
	/*collate conflicting list*/

	/*Edit to return a list of objects instead*/
	json.NewEncoder(w).Encode(value)

}

func CreateServer(port int) *http.Server {
	fmt.Println("Setting up node at port", port)
	var err error

	// Open database
	db, err = bolt.ConnectDB(port)
	if err != nil {
		log.Fatal("Error connecting to Bolt database:", err)
	}
	// defer db.DB.Close()

	// Create bucket
	err = db.CreateBucket("testBucket")
	if err != nil {
		log.Fatal("Error creating bucket:", err)
	}

	/*creates a new instance of a mux router*/
	myRouter := mux.NewRouter().StrictSlash(true)

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