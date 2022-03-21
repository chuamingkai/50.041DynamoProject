package nodes

import (
	"bytes"
	"encoding/gob"
	"log"

	"github.com/chuamingkai/50.041DynamoProject/internal/models"
	rp "github.com/chuamingkai/50.041DynamoProject/pkg/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const bucketName string = "testBucket" // TODO: Change bucket name

func getReplica(key string) (models.Object, bool, error) {
	replicaKey := "replica_" + key
	replica, err := db.Get(bucketName, replicaKey)
	if err != nil {
		return replica, false, err
	} else {
		return replica, true, nil
	}
}

func putReplica(newReplica models.Object) error {
	// Get old value to compare versions
	replicaKey := "replica_" + newReplica.Key
	currentReplica, found, err := getReplica(replicaKey)
	if err != nil {
		return err
	}

	if !found || compareVectorClocks(currentReplica.VC, newReplica.VC) {
		newReplica.Key = replicaKey
		err := db.Put(bucketName, newReplica)
		if err != nil {
			return err
		}
	}
	return nil
}

// TODO: Implement comparison of vector clocks
func compareVectorClocks(currentVC, incomingVC map[string]uint64) bool {
	return true
}

func GetReplica(req *rp.GetRepRequest) (*rp.GetRepResponse, error) {
	stringKey := string(req.Key)
	log.Printf("Getting replica for key %v from data store...\n", stringKey)

	replica, found, err := getReplica(stringKey)

	if err != nil {
		return nil, err
	} else if !found {
		return nil, status.Error(codes.NotFound, "Bolt DB key not found")
	} else {
		var replicaBuffer bytes.Buffer
		err = gob.NewEncoder(&replicaBuffer).Encode(replica)
		if err != nil {
			return nil, err
		}

		replicaBytes := replicaBuffer.Bytes()
		return &rp.GetRepResponse{Data: replicaBytes}, nil
	}
}

func PutReplica(req *rp.PutRepRequest) (*rp.PutRepResponse, error) {
	stringKey := string(req.Key)
	log.Printf("Saving replica for key %v into data store...\n", stringKey)

	var replicaBuffer bytes.Buffer
	var replicaObject models.Object
	err := gob.NewDecoder(&replicaBuffer).Decode(&replicaObject)
	if err != nil {
		return nil, err
	}

	err = putReplica(replicaObject)
	if err != nil {
		return nil, err
	}

	return &rp.PutRepResponse{IsDone: true}, nil
}