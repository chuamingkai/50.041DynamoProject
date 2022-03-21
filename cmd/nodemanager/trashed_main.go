

import (
	"fmt"
	"log"
	//"github.com/gorilla/mux"
	//"sync"
	//"os"
	//"bufio"
	"net/http"
	//"net/http/httptest"
	//"testing"
	"io/ioutil"
	"io"

	"github.com/chuamingkai/50.041DynamoProject/internal/bolt"
	"github.com/chuamingkai/50.041DynamoProject/internal/models"
	"github.com/chuamingkai/50.041DynamoProject/pkg/consistenthashing"
)


type Response struct {
    Status     string // e.g. "200 OK"
    StatusCode int    // e.g. 200
    Proto      string // e.g. "HTTP/1.0"
    ProtoMajor int    // e.g. 1
    ProtoMinor int    // e.g. 0
 
    // response headers
    Header http.Header
    // response body
    Body io.ReadCloser
    // request that was sent to obtain the response
    Request *http.Request
}

func ProcessRequest(node string, ){

}

 /*
func GetFromClient(w http.ResponseWriter, myRouter *http.Request){
	resp, err := http.Get("http://localhost:8000/client")
	if err != nil {
      log.Fatalln(err)
	}
	fmt.Println(resp)

	//We Read the response body on the line below.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
	   log.Fatalln(err)
	}
 //Convert the body to type string
	sb := string(body)
	log.Printf(sb)
 
}
*/




/*
// create server for manager, fixed PORT=8000
func CreateManager(port int) *http.Server {

	fmt.Println("Starting manager @ 8000 <3")
	//creates a new instance of a mux router
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/client", GetFromClient).Methods("GET")

	server := http.Server{
		Addr:    fmt.Sprintf(":%v", port), //:{port}
		Handler: myRouter,
	}
	return &server

}
*/

func main() {
	//var DefaultClient = &http.Client{}

	res, err := http.Get("http://localhost:8000/client")
	//check for response error
	if err != nil {
      log.Fatalln(err)
	}

	//We Read the response body on the line below.
	data, _ := ioutil.ReadAll(res.Body)
	// close response Body
	res.Body.Close()

	//print data as strings
	fmt.Printf("%s\n", (data))

	/*
	// to block to prevent exit of manager
	//exit := make(chan bool)
	go func() {
		server := CreateManager(8000)
		fmt.Println(server.ListenAndServe())
	}()

	//<- exit

	resp, err := http.Get("http://localhost:8000/client")
	if err != nil {
      log.Fatalln(err)
	}
	fmt.Println(resp)

	//We Read the response body on the line below.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
	   log.Fatalln(err)
	}
 //Convert the body to type string
	sb := string(body)
	log.Printf(sb)
	*/


	var ring = consistenthash.NewRing()
	fmt.Println(ring)

	
	id := 55 // TODO: Dynamically assign node IDs
	
	testEntry := models.Object{
		IC: "S1234567A",
		GeoLoc: "23:23:23:23 NW",
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












//just for server testing!
/*
func testServerHandler(t *testing.T){
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
    // pass 'nil' as the third parameter.
    req, err := http.NewRequest("POST", "/client", nil)
    if err != nil {
        t.Fatal(err)
    }

    // We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(PostToServer)

    // Our handlers satisfy http.Handler, so we can call their ServeHTTP method
    // directly and pass in our Request and ResponseRecorder.
    handler.ServeHTTP(rr, req)

    // Check the status code is what we expect.
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v",
            status, http.StatusOK)
    }

    // Check the response body is what we expect.
    expected := `{"alive": true}`
    if rr.Body.String() != expected {
        t.Errorf("handler returned unexpected body: got %v want %v",
            rr.Body.String(), expected)
    }
}
*/