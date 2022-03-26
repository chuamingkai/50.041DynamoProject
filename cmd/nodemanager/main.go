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

<<<<<<< HEAD
// open ports 9000-9100 for client 
=======
// open ports 9000-9100
>>>>>>> 7123d86285d92c11d81f0aae7158692c23c81c14

// let's declare a global Keys array
// that we can then populate in our main function
// to simulate a database
var Keys []Key


// HTPP:{"Key":"asd", "Value":"sdf"}
type Key struct {
	Id      string `json:"Id"`
    Title string `json:"Title"`
    Desc string `json:"desc"`
    Content string `json:"content"`
}




func homePage(w http.ResponseWriter, r *http.Request){
    fmt.Fprintf(w, "Client landing!")
    fmt.Println("Endpoint Hit: nodemanager http server")
}

func handleRequests() {
    // creates a new instance of a mux router
    myRouter := mux.NewRouter().StrictSlash(true)
    // replace http.HandleFunc with myRouter.HandleFunc
    myRouter.HandleFunc("/", homePage)
    myRouter.HandleFunc("/all", returnAllKeys)
    // finally, instead of passing in nil, we want
    // to pass in our newly created router as the second
    // argument
	// NOTE: Ordering is important here! This has to be defined before
    // the other `/key` endpoint. 
    myRouter.HandleFunc("/key", createNewKey).Methods("POST")
    myRouter.HandleFunc("/key/{id}", returnSingleKey)
    log.Fatal(http.ListenAndServe(":8000", myRouter))
}

func returnSingleKey(w http.ResponseWriter, r *http.Request){
    vars := mux.Vars(r)
    key := vars["id"]

    // Loop over all of our Articles
    // if the article.Id equals the key we pass in
    // return the article encoded as JSON
    for _, article := range Keys {
        if article.Id == key {
            json.NewEncoder(w).Encode(article)
        }
    }
}


// encode keys to arr->json str
func returnAllKeys(w http.ResponseWriter, r *http.Request){
    fmt.Println("Endpoint Hit: returnAllKeys")
    json.NewEncoder(w).Encode(Keys)
}


func createNewKey(w http.ResponseWriter, r *http.Request) {
    // get the body of our POST request
    // unmarshal this into a new Article struct
    // append this to our Articles array.    
    reqBody, _ := ioutil.ReadAll(r.Body)
    var article Key 
    json.Unmarshal(reqBody, &article)
    // update our global Articles array to include
    // our new Article
    Keys = append(Keys, article)

    json.NewEncoder(w).Encode(article)
}

func deleteKey(w http.ResponseWriter, r *http.Request) {
    // once again, we will need to parse the path parameters
    vars := mux.Vars(r)
    // we will need to extract the `id` of the article we
    // wish to delete
    id := vars["id"]

    // we then need to loop through all our articles
    for index, key := range Keys {
        // if our id path parameter matches one of our
        // articles
        if key.Id == id {
            // updates our Articles array to remove the 
            // article
            Keys = append(Keys[:index], Keys[index+1:]...)
        }
    }

}


func main() {
 
	//populate with dummy data
	Keys = []Key{
        Key{Id: "1", Title: "Hello", Desc: "Article Description", Content: "Article Content"},
        Key{Id: "2", Title: "Hello 2", Desc: "Article Description", Content: "Article Content"},
    }
<<<<<<< HEAD

=======
>>>>>>> 7123d86285d92c11d81f0aae7158692c23c81c14

	handleRequests()

	//---------------------------------------
	var ring = consistenthash.NewRing()
	fmt.Println(ring)

	
	// id := 55 // TODO: Dynamically assign node IDs
	
<<<<<<< HEAD
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
=======
	// testEntry := models.Object{
	// 	IC: "S1234567A",
	// 	GeoLoc: "23:23:23:23 NW",
	// }

	// // Open database
	// db, err := bolt.ConnectDB(id)
	// if err != nil {
	// 	log.Fatalf("Error opening database: %s", err)
	// }
	// defer db.DB.Close()

	// // Create bucket
	// err = db.CreateBucket("testBucket")
	// if err != nil {
	// 	log.Fatalf("Error creating bucket: %s", err)
	// }

	// // Insert test value into bucket
	// err = db.Put("testBucket", testEntry)
	// if err != nil {
	// 	log.Fatalf("Error inserting into bucket: %s", err)
	// }

	// // Read from bucket
	// value := db.Get("testBucket", testEntry.IC)
	// fmt.Printf("Value at key %s: %s", testEntry.IC, value)
}
>>>>>>> 7123d86285d92c11d81f0aae7158692c23c81c14
