package nodes

import (
	"bytes"
	"context"
	"encoding/gob"
	"log"

	"github.com/chuamingkai/50.041DynamoProject/internal/models"
	pb "github.com/chuamingkai/50.041DynamoProject/pkg/internalcomm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *nodesServer) doGetReplica(bucketName, key string) (models.Object, bool, error) {
	replica, err := s.boltDB.Get(bucketName, key)
	if err != nil {
		return replica, false, err
	} else {
		return replica, true, nil
	}
}

func (s *nodesServer) doPutReplica(bucketName, key string, newReplica models.Object) error {
	// Get old value to compare versions
	currentReplica, found, err := s.doGetReplica(bucketName, key)
	if err != nil {
		return err
	}

	if !found || compareVectorClocks(currentReplica.VC, newReplica.VC) {
		err = s.boltDB.Put(bucketName, newReplica)
	}
	return err
}

// GetReplica issued from server responsible for the get operation
func (s *nodesServer) GetReplica(ctx context.Context, req *pb.GetRepRequest) (*pb.GetRepResponse, error) {
	log.Printf("Getting replica for key: {%v} from local db\n", string(req.Key))

	replica, found, err := s.doGetReplica(req.BucketName, req.Key)

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
		return &pb.GetRepResponse{Data: replicaBytes}, nil
	}
}

// PutReplica issued from server responsible for the get operation
func (s *nodesServer) PutReplica(ctx context.Context, req *pb.PutRepRequest) (*pb.PutRepResponse, error) {
	replicaBuffer := bytes.NewBuffer(req.Data)
	var replicaObject models.Object
	err := gob.NewDecoder(replicaBuffer).Decode(&replicaObject)
	if err != nil {
		return nil, err
	}
	log.Printf("Putting replica key: %s, val: %v to local db", req.Key, replicaObject)

	err = s.doPutReplica(req.BucketName, req.Key, replicaObject)
	if err != nil {
		return nil, err
	}

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