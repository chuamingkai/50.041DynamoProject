package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/DistributedClocks/GoVector/govec/vclock"
	"github.com/gorilla/mux"
)

type Object struct {
	IC     string            `json:"ic"`
	GeoLoc string            `json:"geoloc"`
	VC     map[string]uint64 `json:",omitempty"`
}

/*Temporary variable*/
var Entries []Object

/*wrapper function to increase vector clock value of a particular index by 1*/
/*i.e. vc1={"A":1}, update("B",vc1)={"A":1,"B":1}*/
func update(index string, orig vclock.VClock) vclock.VClock {
	updated := orig.Copy()
	old, check := updated.FindTicks(index)
	if !check {
		updated.Set(index, 1)
	} else {
		updated.Set(index, old+1)
	}
	return updated.Copy()
}

func getPreferenceList(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Preference List")

}

func returnAllEntries(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(Entries)
}

func createNewEntry(w http.ResponseWriter, r *http.Request) {

	reqBody, _ := ioutil.ReadAll(r.Body)
	var newentry Object
	json.Unmarshal(reqBody, &newentry)

	Entries = append(Entries, newentry)

	json.NewEncoder(w).Encode(newentry)
}

func returnSingleEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["ic"]

	for _, entry := range Entries {
		if entry.IC == key {
			json.NewEncoder(w).Encode(entry)
		}
	}
}

func handleRequests() {
	/*creates a new instance of a mux router*/
	myRouter := mux.NewRouter().StrictSlash(true)

	/*get preference list*/
	myRouter.HandleFunc("/", getPreferenceList)

	/*write new entry*/

	myRouter.HandleFunc("/data", createNewEntry).Methods("POST")

	/*get all entries*/
	myRouter.HandleFunc("/data", returnAllEntries)

	/*return single entry*/
	myRouter.HandleFunc("/data/{ic}", returnSingleEntry)

	log.Fatal(http.ListenAndServe(":8080", myRouter))
}

func main() {
	vc1 := vclock.New()
	vc1.Set("A", 1)
	vc2 := update("B", vc1)
	Entries = []Object{
		Object{IC: "A12345678A", GeoLoc: "somewhere2", VC: vc1.GetMap()},
		Object{IC: "S12345678A", GeoLoc: "somewhere1", VC: vc2.GetMap()},
	}

	handleRequests()

}
