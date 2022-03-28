package nodes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/chuamingkai/50.041DynamoProject/internal/models"
	consistenthash "github.com/chuamingkai/50.041DynamoProject/pkg/consistenthashing"
	pb "github.com/chuamingkai/50.041DynamoProject/pkg/internalcomm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type GetRepMessage struct {
	peerId uint64
	repObject models.Object
	err error
}

type GetRepResult struct {
	Replicas []models.Object `json:"replicas"`
	HasVersionConflict bool `json:"hasVersionConflict"`
}

type PutRepMessage struct {
	peerId uint64
	putSuccess bool
	err error
}

type PutRepResult struct {
	Success bool `json:"success"`
	Err error `json:"error"`
}

func (s *nodesServer) serverGetReplicas(ctx context.Context, req *pb.GetRepRequest) (*GetRepResult, error) {
	log.Printf("Received GET request from clients for key %s\n", req.Key)

	// Get preference list for the key
	prefList := s.ring.GetPreferenceList(req.Key)

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

			log.Printf("Getting replica from node %v.\n", vn.NodeId)
			if vn.NodeId == uint64(s.nodeId) {
				getRepRes, err = s.GetReplica(ctxRep, &getRepReq)
			} else {
				// var peer pb.ReplicationClient
				peerAddr := fmt.Sprintf("localhost:%v", vn.NodeId)
				dialOptions := grpc.WithTransportCredentials(insecure.NewCredentials())
				conn, err := grpc.Dial(peerAddr, dialOptions)
				
				if err == nil {
					peer := pb.NewReplicationClient(conn)
					getRepRes, err = peer.GetReplica(ctxRep, &getRepReq)
					if err != nil {
						log.Printf("Error getting replica from node %v: %v\n", vn.NodeId, err.Error())
					}
				} else {
					log.Printf("Error contacting node %v: %v\n", vn.NodeId, err)
					// notifyChan <- GetRepMessage{
					// 	peerId: vn.NodeId,
					// 	err: err,
					// }
					// return
				}
				defer conn.Close()
			}

			if err != nil {
				notifyChan <- GetRepMessage{
					peerId: vn.NodeId,
					err: err,
				}
				return
			}

			var repObject models.Object
			if getRepRes.Data != nil {
				decodeErr := json.Unmarshal(bytes.Trim(getRepRes.Data, "\x00"), &repObject)
				if decodeErr != nil {
					notifyChan <- GetRepMessage{
						peerId: vn.NodeId,
						err: decodeErr,
					}
				} else {
					notifyChan <- GetRepMessage{
						peerId: vn.NodeId,
						repObject: repObject,
					}
				}
			} else {
				notifyChan <- GetRepMessage{
					peerId: vn.NodeId,
					err: fmt.Errorf("failed to get replica from node %v", vn.NodeId),
				}
			}
			
		}(virtualNode, notifyChan)
	}

	var replicas []models.Object

	successCount := 0
	errorMsg := "ERROR: "
	for i := 0; i < len(prefList); i++ {
		select {
		case notifyMsg := <- notifyChan:
			if notifyMsg.err == nil {
				log.Printf("Received GET replica response from node %v for key {%v}: val {%v}", notifyMsg.peerId, req.Key, notifyMsg.repObject)
				successCount++
				replicas = append(replicas, notifyMsg.repObject)
			} else if e, ok := status.FromError(notifyMsg.err); ok && e.Code() == codes.NotFound {
				log.Printf("Received GET replica response from node %v for key {%v}: key not found", notifyMsg.peerId, req.Key)
				successCount++
			} else {
				errorMsg = errorMsg + notifyMsg.err.Error() + ";"
				log.Printf("Received GET replica response from node %v for key {%v}: Error occurred: %v", notifyMsg.peerId, req.Key, notifyMsg.err.Error())
			}
		case <- ctxRep.Done():
		}
		if ctxRep.Err() != nil || successCount == consistenthash.REPLICATION_FACTOR {
			break
		}
	}

	log.Printf("Finished getting replicas for key %v, received %v successful response(s).\n", req.Key, successCount)

		
	if successCount == 0 { // No responses received -> servers in preference list are down
		return nil, errors.New(errorMsg)
	} else if len(replicas) == 0 { // No replicas received -> key does not exist in database
		return &GetRepResult{Replicas: replicas, HasVersionConflict: false}, errors.New("key does not exist")
	} else if allSame(replicas) { // No version conflicts among replicas
		return &GetRepResult{Replicas: replicas, HasVersionConflict: false}, nil
	} else { // Versioning conflicts found
		return &GetRepResult{Replicas: replicas, HasVersionConflict: true}, nil
	}
}

func (s *nodesServer) serverPutReplicas(ctx context.Context, req *pb.PutRepRequest) *PutRepResult {
	var reqObject models.Object
	if err := json.Unmarshal(req.Data, &reqObject); err != nil{
		return &PutRepResult{Success: false, Err: err}
	}
	log.Printf("Received PUT request from clients for key %s, value %v\n", req.Key, reqObject)

	// Get preference list for the key
	prefList := s.ring.GetPreferenceList(req.Key)

	// Create context to get replicas
	clientDeadline := time.Now().Add(5 * time.Second)
	ctxRep, cancel := context.WithDeadline(ctx, clientDeadline)
	defer cancel()

	// Send GetRep request to all servers in preference list
	notifyChan := make(chan PutRepMessage, consistenthash.REPLICATION_FACTOR)	
	putRepReq := &pb.PutRepRequest{
		Key: req.Key,
		Data: req.Data,
		BucketName: req.BucketName,
	}
	for _, virtualNode := range prefList {
		var putRepRes *pb.PutRepResponse
		var err error

		go func(vn consistenthash.VirtualNode, notifyChan chan PutRepMessage) {
			if vn.NodeId == uint64(s.nodeId) {
				putRepRes, err = s.PutReplica(ctxRep, putRepReq)
			} else {
				peerAddr := fmt.Sprintf("localhost:%v", vn.NodeId)
				dialOptions := grpc.WithTransportCredentials(insecure.NewCredentials())
				conn, err := grpc.Dial(peerAddr, dialOptions)
				
				if err == nil {
					peer := pb.NewReplicationClient(conn)
					putRepRes, err = peer.PutReplica(ctxRep, putRepReq)
					if err != nil {
						log.Printf("Error putting replica in node %v: %v\n", vn.NodeId, err.Error())
					}
				} else {
					log.Printf("Error contacting node %v: %v\n", vn.NodeId, err)
				}
				defer conn.Close()
			}
			notifyChan <- PutRepMessage{
				peerId: vn.NodeId,
				putSuccess: putRepRes.IsDone,
				err: err,
			}
		}(virtualNode, notifyChan)
	}

	successCount := 0
	errorMsg := "ERROR MESSAGE: "
	for i := 0; i < len(prefList); i++ {
		select {
		case notifyMsg := <- notifyChan:
			if notifyMsg.err == nil {
				log.Printf("Received PUT replica response from server %v for key {%v} val {%v}: SUCCESS", notifyMsg.peerId, req.Key, reqObject)
				successCount++
			} else {
				errorMsg = errorMsg + notifyMsg.err.Error() + ";"
				log.Printf("Received PUT replica response from server %v for key {%v} val {%v}: ERROR occured: %v\n", notifyMsg.peerId, req.Key, reqObject, notifyMsg.err.Error())
			}
		case <- ctxRep.Done():
		}
		if ctxRep.Err() != nil || successCount == consistenthash.REPLICATION_FACTOR {
			break
		}
	}

	log.Printf("Finished PUTTING replica for key {%v} val {%v}, received {%v} response", req.Key, reqObject, successCount)
	if successCount == 0 { // All servers failed to save replica
		return &PutRepResult{Success: false, Err: errors.New(errorMsg)}
	} else {
		// TODO: Add case to handle partial success (not all servers in preference list successfuly saved replica)
		return &PutRepResult{Success: true, Err: nil}
	}
} 

func allSame(replicas []models.Object) bool {
	for i := 1; i < len(replicas); i++ {
		// TODO: Figure out how to check for versioning conflicts
		if replicas[0].Value != replicas[i].Value {
			return false
		}
	}
	return true
}