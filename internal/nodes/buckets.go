package nodes

import (
	"encoding/json"
	"net/http"

	"github.com/chuamingkai/50.041DynamoProject/internal/models"
	"github.com/gorilla/mux"
)

// Get all objects inside a bucket
func (s *nodesServer) doGetAllBucketObjects(w http.ResponseWriter, r *http.Request) {
	pathVars := mux.Vars(r)
	var ok bool
	var bucketName string
	if bucketName, ok = pathVars["bucketName"]; !ok {
		http.Error(w, "Missing bucket name", http.StatusBadRequest)
		return
	}

	bucketObjectsBytes, err := s.boltDB.GetAllObjects(bucketName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var bucketObjects []models.Object
	for _, objBytes := range bucketObjectsBytes {
		var obj models.Object
		if err := json.Unmarshal(objBytes, &obj); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		bucketObjects = append(bucketObjects, obj)
	}

	json.NewEncoder(w).Encode(bucketObjects)
}

// Get all buckets inside the database
func (s *nodesServer) getAllBuckets(w http.ResponseWriter, r *http.Request) {
	bucketNames, err := s.boltDB.GetAllBuckets()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		json.NewEncoder(w).Encode(bucketNames)
	}
}

// Delete key from bucket
func (s *nodesServer) deleteKey(w http.ResponseWriter, r *http.Request) {
	pathVars := mux.Vars(r)
	var ok bool
	var bucketName, key string

	if bucketName, ok = pathVars["bucketName"]; !ok {
		http.Error(w, "Missing bucket name", http.StatusBadRequest)
		return
	}

	if key, ok = pathVars["key"]; !ok {
		http.Error(w, "Missing key to delete", http.StatusBadRequest)
		return
	}

	if err := s.boltDB.DeleteKey(bucketName, key); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}