package nodes

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/chuamingkai/50.041DynamoProject/internal/models"
	pb "github.com/chuamingkai/50.041DynamoProject/pkg/internalcomm"
	"github.com/chuamingkai/50.041DynamoProject/pkg/vectorclock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *nodesServer) doGetReplica(bucketName, key string) ([]byte, bool, error) {
	// Check if bucket exists
	if !s.boltDB.BucketExists(bucketName) {
		return nil, false, fmt.Errorf("bucket does not exist on node %v", s.nodeId)
	}

	replica, err := s.boltDB.Get(bucketName, key)
	if err != nil {
		return replica, false, err
	} else {
		return replica, true, nil
	}
}

func (s *nodesServer) doPutReplica(bucketName, key string, data []byte, senderId int64) error {
	if !s.boltDB.BucketExists(bucketName) {
		// Create bucket if it does not exist
		if err := s.boltDB.CreateBucket(bucketName); err != nil {
			return err
		}
		
	}

	var incomingRepObject models.Object
	if err := json.Unmarshal(data, &incomingRepObject); err != nil {
		return err
	}
	currentTime := time.Now()
	nodename := strconv.Itoa(int(senderId))

	// TODO: Vector clock comparison to avoid versioning conflicts
	existingRep, err := s.boltDB.Get(bucketName, key)
	if err != nil {
		return err
	} else if existingRep != nil {
		var existingRepObject models.Object
		if err := json.Unmarshal(existingRep, &existingRepObject); err != nil {
			return err
		}
		incomingRepObject.CreatedOn = existingRepObject.CreatedOn
		incomingRepObject.LastModifiedOn = currentTime
		incomingRepObject.VC = vectorclock.UpdateRecv(nodename, existingRepObject.VC)

		incomingRepBytes, err := json.Marshal(incomingRepObject)
		if err != nil {
			return err
		}
		
		return s.boltDB.Put(bucketName, key, incomingRepBytes)
	} else {
		incomingRepObject.CreatedOn = currentTime
		incomingRepObject.LastModifiedOn = currentTime
		incomingRepObject.VC = vectorclock.UpdateRecv(nodename, map[string]uint64{nodename: 0})

		incomingRepBytes, err := json.Marshal(incomingRepObject)
		if err != nil {
			return err
		}
		return s.boltDB.Put(bucketName, key, incomingRepBytes)
	}
}


// GetReplica issued from server responsible for the get operation
func (s *nodesServer) GetReplica(ctx context.Context, req *pb.GetRepRequest) (*pb.GetRepResponse, error) {
	log.Printf("Received GET request for replica for key '%v'\n", req.Key)

	// TODO: Figure out what to do when the replica's key is not found
	replica, found, err := s.doGetReplica(req.BucketName, req.Key)
	// log.Printf("Replica for key '%v' found\n", req.Key)
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

	err := s.doPutReplica(req.BucketName, req.Key, req.Data, req.SenderId)
	if err != nil {
		return &pb.PutRepResponse{IsDone: false}, err
	}

	log.Printf("Successfully PUT replica for key '%s', val: %s\n", req.Key, req.Data)
	return &pb.PutRepResponse{IsDone: true}, nil
}