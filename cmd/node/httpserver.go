package main

import (
	"fmt"
	"sync"

	"github.com/chuamingkai/50.041DynamoProject/internal/nodes"
)

/*wrapper function to increase vector clock value of a particular index by 1*/
/*i.e. vc1={"A":1}, update("B",vc1)={"A":1,"B":1}*/
// func updaterecv(index string, orig map[string]uint64) map[string]uint64 {
// 	updated := vclock.New().CopyFromMap(orig)
// 	old, check := updated.FindTicks(index)
// 	if !check {
// 		updated.Set(index, 1)
// 	} else {
// 		updated.Set(index, old+1)
// 	}
// 	return updated.GetMap()
// }

// // TODO: Move these into a different folder, modify the db writing code
// func createSingleEntry(w http.ResponseWriter, r *http.Request) {

// 	/*To add check for if node handles key*/
// 	nodename := strings.Trim(r.Host, "localhost:")
// 	id, _ := strconv.Atoi(nodename)

// 	// Open database
// 	db, err := bolt.ConnectDB(id)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer db.DB.Close()

// 	// Create bucket
// 	db.CreateBucket("testBucket")

// 	reqBody, _ := ioutil.ReadAll(r.Body)
// 	var data map[string]string
// 	json.Unmarshal(reqBody, &data)
// 	fmt.Println(data)

// 	var newEntry models.Object
// 	for k, v := range data {
// 		fmt.Printf("Key: %s, Value: %s\n", k, v)
// 		newEntry.Key = k
// 		newEntry.Value = v
// 	}

// 	// var newentry models.Object
// 	// json.Unmarshal(reqBody, &newentry)
// 	if newEntry.VC != nil {
// 		newEntry.VC = vectorclock.UpdateRecv(nodename, newEntry.VC)
// 	} else {
// 		newEntry.VC = vectorclock.UpdateRecv(nodename, map[string]uint64{nodename: 0})
// 	}

// 	/*To add ignore if new updated vector clock is outdated?*/

// 	// Insert test value into bucket
// 	err = db.Put("testBucket", newEntry)
// 	if err != nil {
// 		log.Fatalf("Error inserting into bucket: %s", err)
// 	}

// 	/*To add gRPC contact*/
// 	/*record first response-> writeCoordinator*/

// 	/*Edit to return a list of objects instead*/
// 	json.NewEncoder(w).Encode(newEntry)
// }

// func returnSingleEntry(w http.ResponseWriter, r *http.Request) {

// 	/*To add check for if node handles key*/
// 	nodename := strings.Trim(r.Host, "localhost:")
// 	id, _ := strconv.Atoi(nodename)

// 	vars := mux.Vars(r)
// 	key := vars["ic"]
// 	fmt.Println(key)

// 	// Open database
// 	db, err := bolt.ConnectDB(id)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer db.DB.Close()

// 	// Read from bucket
// 	value := db.Get("testBucket", key)
// 	fmt.Println(value)

// 	/*To add gRPC contact*/
// 	/*collate conflicting list*/

// 	/*Edit to return a list of objects instead*/
// 	json.NewEncoder(w).Encode(value)

// }

// func createServer(port int) *http.Server {
// 	fmt.Println("Node running at port", port)
// 	/*creates a new instance of a mux router*/
// 	myRouter := mux.NewRouter().StrictSlash(true)

// 	/*write new entry*/

// 	myRouter.HandleFunc("/data", createSingleEntry).Methods("POST")

// 	/*return single entry*/
// 	myRouter.HandleFunc("/data/{ic}", returnSingleEntry)

// 	server := http.Server{
// 		Addr:    fmt.Sprintf(":%v", port), //:{port}
// 		Handler: myRouter,
// 	}
// 	return &server

// }

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
		server := nodes.CreateServer(9000)
		fmt.Println(server.ListenAndServe())
		wg.Done()
	}()

	go func() {
		server := nodes.CreateServer(9022)
		fmt.Println(server.ListenAndServe())
		wg.Done()
	}()

	wg.Wait()
}
