package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/chuamingkai/50.041DynamoProject/internal/bolt"
	"github.com/chuamingkai/50.041DynamoProject/internal/models"

	"github.com/DistributedClocks/GoVector/govec/vclock"
	"github.com/gorilla/mux"
)

/*wrapper function to increase vector clock value of a particular index by 1*/
/*i.e. vc1={"A":1}, update("B",vc1)={"A":1,"B":1}*/
func updaterecv(index string, orig map[string]uint64) map[string]uint64 {
	updated := vclock.New().CopyFromMap(orig)
	old, check := updated.FindTicks(index)
	if !check {
		updated.Set(index, 1)
	} else {
		updated.Set(index, old+1)
	}
	return updated.GetMap()
}

func createSingleEntry(w http.ResponseWriter, r *http.Request) {

	/*To add check for if node handles key*/
	nodename := strings.Trim(r.Host, "localhost:")
	id, err := strconv.Atoi(nodename)

	// Open database
	db, err := bolt.ConnectDB(id)
	if err != nil {
		log.Fatal(err)
	}
	defer db.DB.Close()

	reqBody, _ := ioutil.ReadAll(r.Body)
	var newentry models.Object
	json.Unmarshal(reqBody, &newentry)
	if newentry.VC != nil {
		newentry.VC = updaterecv(nodename, newentry.VC)
	} else {
		newentry.VC = updaterecv(nodename, map[string]uint64{nodename: 0})
	}

	/*To add ignore if new updated vector clock is outdated?*/

	// Insert test value into bucket
	err = db.Put("testBucket", newentry)
	if err != nil {
		log.Fatalf("Error inserting into bucket: %s", err)
	}

	/*To add gRPC contact*/
	/*record first response-> writeCoordinator*/

	/*Edit to return a list of objects instead*/
	json.NewEncoder(w).Encode(newentry)

}

func returnSingleEntry(w http.ResponseWriter, r *http.Request) {

	/*To add check for if node handles key*/
	nodename := strings.Trim(r.Host, "localhost:")
	id, err := strconv.Atoi(nodename)

	vars := mux.Vars(r)
	key := vars["ic"]
	fmt.Println(key)

	// Open database
	db, err := bolt.ConnectDB(id)
	if err != nil {
		log.Fatal(err)
	}
	defer db.DB.Close()

	// Read from bucket
	value := db.Get("testBucket", key)
	fmt.Printf("Value at key %s: %s", key, value.GeoLoc)

	/*To add gRPC contact*/
	/*collate conflicting list*/

	/*Edit to return a list of objects instead*/
	json.NewEncoder(w).Encode(value)

}

func createServer(port int) *http.Server {
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

func main() {
	wg := new(sync.WaitGroup)
	wg.Add(2)
	/*
		//If database not created before
		go func() {
			db, err := bolt.ConnectDB(9000)
			if err != nil {
				log.Fatal(err)
			}
			defer db.DB.Close()

			// Create bucket
			err = db.CreateBucket("testBucket")
			if err != nil {
				log.Fatalf("Error creating bucket: %s", err)
			}
		}()
	*/

	go func() {
		server := createServer(9000)
		fmt.Println(server.ListenAndServe())
		wg.Done()
	}()

	go func() {
		server := createServer(9022)
		fmt.Println(server.ListenAndServe())
		wg.Done()
	}()

	wg.Wait()
}
