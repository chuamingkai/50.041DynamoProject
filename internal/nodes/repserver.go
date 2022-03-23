package nodes

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	consistenthash "github.com/chuamingkai/50.041DynamoProject/pkg/consistenthashing"
	pb "github.com/chuamingkai/50.041DynamoProject/pkg/internalcomm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetRepMessage struct {
	peerId uint64
	repObject []byte
	err error
}

func (s *nodesServer) Get(ctx context.Context, req *pb.GetRepRequest) (*pb.GetRepResponse, error) {
	log.Printf("Received GET request from clients for key %s\n", req.Key)

	// Get preference list for the key
	prefList := s.ring.GetPreferenceList(req.Key)

	// View preference list
	// for _, vn := range prefList {
	// 	log.Printf("Peer for node %v: %v", s.nodeId, vn)
	// }

	// Create context to get replicas
	clientDeadline := time.Now().Add(5 * time.Second)
	ctxRep, cancel := context.WithDeadline(ctx, clientDeadline)
	defer cancel()

	// Send GetRep request to all servers in preference list
	notifyChan := make(chan GetRepMessage, consistenthash.REPLICATION_FACTOR)
	for _, virtualNode := range prefList {
		getRepReq := pb.GetRepRequest{
			BucketName: req.BucketName,
			Key: req.Key,
		}

		// Send GetRep request to virtual node
		go func(vn consistenthash.VirtualNode, notifyChan chan GetRepMessage) {
			var getRepRes *pb.GetRepResponse
			var err error

			if vn.NodeId == uint64(s.nodeId) {
				getRepRes, err = s.GetReplica(ctxRep, &getRepReq)
			} else {
				// var peer pb.ReplicationClient
				peerAddr := fmt.Sprintf("localhost:%v", vn.NodeId)
				conn, err := grpc.Dial(peerAddr)
				if err == nil {
					peer := pb.NewReplicationClient(conn)
					getRepRes, err = peer.GetReplica(ctxRep, &getRepReq)
				}
				defer conn.Close()
			}

			if err != nil {
				notifyChan <- GetRepMessage{
					peerId: vn.NodeId,
					repObject: getRepRes.Data,
					err: nil,
				}
			} else {
				notifyChan <- GetRepMessage{
					peerId: vn.NodeId,
					repObject: nil,
					err: err,
				}
			}
		}(virtualNode, notifyChan)
	}

	var replicas [][]byte

	successCount := 0
	errorMsg := "ERROR: "
	for i := 0; i < len(prefList); i++ {
		select {
		case notifyMsg := <- notifyChan:
			if notifyMsg.err == nil {
				log.Printf("Received GET replica response from node %v for key {%v}: val {%v}", notifyMsg.peerId, req.Key, string(notifyMsg.repObject))
				successCount++
				replicas = append(replicas, notifyMsg.repObject)
			} else if e, ok := status.FromError(notifyMsg.err); ok && e.Code() == codes.NotFound {
				log.Printf("Received GET replica response from server %v for key {%v}: key not found", notifyMsg.peerId, req.Key)
				successCount++
			} else {
				errorMsg = errorMsg + notifyMsg.err.Error() + ";"
				log.Printf("Received GET repica response from server %v for key {%v}: occured error: %v", notifyMsg.peerId, req.Key, notifyMsg.err.Error())
			}
		case <- ctxRep.Done():
		}
		if ctxRep.Err() != nil || successCount == consistenthash.REPLICATION_FACTOR {
			break
		}
	}

	log.Printf("Finished getting replicas for key %v, received %v response(s).\n", req.Key, successCount)

		
	if successCount == 0 { // No responses received -> servers in preference list are down
		return nil, errors.New(errorMsg)
	} else if len(replicas) == 0 { // No replicas received -> key does not exist in database
		return &pb.GetRepResponse{Data: nil, HasVersionConflict: false}, nil
	} else if allSame(replicas) { // No version conflicts among replicas
		return &pb.GetRepResponse{Data: replicas[0], HasVersionConflict: false}, nil
	} else { // Versioning conflicts found
		// TODO: Return list of all replicas
		return &pb.GetRepResponse{Data: replicas[1], HasVersionConflict: true}, nil
	}
}

// TODO: Add Put() method

func allSame(replicas [][]byte) bool {
	for i := 1; i < len(replicas); i++ {
		if !bytes.Equal(replicas[0], replicas[i]) {
			return false
		}
	}
	return true
}