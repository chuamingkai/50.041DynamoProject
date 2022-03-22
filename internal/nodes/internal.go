package nodes

import (
	"bytes"
	"encoding/gob"
	"log"

	"github.com/chuamingkai/50.041DynamoProject/internal/bolt"
	"github.com/chuamingkai/50.041DynamoProject/internal/models"
	pb "github.com/chuamingkai/50.041DynamoProject/pkg/internalcomm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TODO: Remove replicaBucket -> replicas go into the same bucket as non-replicas (non-distinguishable)

type replicaServer struct {
	pb.UnimplementedReplicationServer
	boltDB *bolt.DB // Server's own DB pointer
	replicaBucket string // Name of bucket to store replicas
}

func (s *replicaServer) doGetReplica(key string) (models.Object, bool, error) {
	replicaKey := "replica_" + key
	replica, err := s.boltDB.Get(s.replicaBucket, replicaKey)
	if err != nil {
		return replica, false, err
	} else {
		return replica, true, nil
	}
}

func (s *replicaServer) doPutReplica(newReplica models.Object) error {
	// Get old value to compare versions
	replicaKey := "replica_" + newReplica.Key
	currentReplica, found, err := s.doGetReplica(replicaKey)
	if err != nil {
		return err
	}

	if !found || compareVectorClocks(currentReplica.VC, newReplica.VC) {
		newReplica.Key = replicaKey
		err := s.boltDB.Put(s.replicaBucket, newReplica)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *replicaServer) GetReplica(req *pb.GetRepRequest) (*pb.GetRepResponse, error) {
	stringKey := string(req.Key)
	log.Printf("Getting replica for key %v from data store...\n", stringKey)

	replica, found, err := s.doGetReplica(stringKey)

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

func (s *replicaServer) PutReplica(req *pb.PutRepRequest) (*pb.PutRepResponse, error) {
	stringKey := string(req.Key)
	log.Printf("Saving replica for key %v into data store...\n", stringKey)

	var replicaBuffer bytes.Buffer
	var replicaObject models.Object
	err := gob.NewDecoder(&replicaBuffer).Decode(&replicaObject)
	if err != nil {
		return nil, err
	}

	err = s.doPutReplica(replicaObject)
	if err != nil {
		return nil, err
	}

	return &pb.PutRepResponse{IsDone: true}, nil
}

// TODO: Implement comparison of vector clocks
func compareVectorClocks(currentVC, incomingVC map[string]uint64) bool {
	return true
}

func newReplicaServer(boltDB *bolt.DB) *replicaServer {
	return &replicaServer{
		boltDB: boltDB,
		replicaBucket: "replicas",
	}
}