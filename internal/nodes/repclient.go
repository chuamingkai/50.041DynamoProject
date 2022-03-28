package nodes

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"log"

	pb "github.com/chuamingkai/50.041DynamoProject/pkg/internalcomm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TODO: Add relevant logging messages

func (s *nodesServer) doGetReplica(bucketName, key string) ([]byte, bool, error) {
	// Check if bucket exists
	if !s.boltDB.BucketExists(bucketName) {
		return nil, false, errors.New("Bucket does not exist on node" + string(s.nodeId))
	}

	replica, err := s.boltDB.Get(bucketName, key)
	if err != nil {
		return replica, false, err
	} else {
		return replica, true, nil
	}
}

func (s *nodesServer) doPutReplica(bucketName, key string, data []byte) error {
	if !s.boltDB.BucketExists(bucketName) {
		return errors.New("Bucket does not exist on node" + string(s.nodeId))
	}
	// Get old value to compare versions
	// currentReplicaBytes, found, err := s.doGetReplica(bucketName, key)
	// if err != nil {
	// 	return err
	// }

	// currentReplicaBuffer := bytes.NewBuffer(currentReplicaBytes)
	// var currentReplica models.Object
	// gob.NewDecoder(currentReplicaBuffer).Decode(&currentReplica)

	// if !found || compareVectorClocks(currentReplica.VC, newReplica.VC) {
	// 	err = s.boltDB.Put(bucketName, newReplica)
	// }
	// return err
	// TODO: Vector clock comparison to avoid versioning conflicts
	return s.boltDB.Put(bucketName, key, data)
}


// GetReplica issued from server responsible for the get operation
func (s *nodesServer) GetReplica(ctx context.Context, req *pb.GetRepRequest) (*pb.GetRepResponse, error) {
	log.Printf("Received GET request for replica for key '%v'\n", req.Key)

	// TODO: Figure out what to do when the replica's key is not found
	replica, found, err := s.doGetReplica(req.BucketName, req.Key)
	log.Printf("Replica for key '%v' found: %s\n", req.Key, replica)
	if err != nil {
		return nil, err
	} else if !found {
		return nil, status.Error(codes.NotFound, "Key not found")
	} else {
		var replicaBuffer bytes.Buffer
		err = gob.NewEncoder(&replicaBuffer).Encode(replica)
		if err != nil {
			return nil, err
		}

		return &pb.GetRepResponse{Data: replica}, nil
	}
}

// PutReplica issued from server responsible for the get operation
func (s *nodesServer) PutReplica(ctx context.Context, req *pb.PutRepRequest) (*pb.PutRepResponse, error) {
	log.Printf("Received PUT request for replica for key '%s', val: %s\n", req.Key, req.Data)

	err := s.doPutReplica(req.BucketName, req.Key, req.Data)
	if err != nil {
		return &pb.PutRepResponse{IsDone: false}, err
	}

	log.Printf("Successfully PUT replica for key '%s', val: %s\n", req.Key, req.Data)
	return &pb.PutRepResponse{IsDone: true}, nil
}

// TODO: Implement comparison of vector clocks
func compareVectorClocks(currentVC, incomingVC map[string]uint64) bool {
	return true
}

// func newReplicaServer(boltDB *bolt.DB) *replicaServer {
// 	return &replicaServer{
// 		boltDB: boltDB,
// 	}
// }