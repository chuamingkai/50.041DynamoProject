package nodes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	config "github.com/chuamingkai/50.041DynamoProject/config"
	"github.com/chuamingkai/50.041DynamoProject/internal/models"
	consistenthash "github.com/chuamingkai/50.041DynamoProject/pkg/consistenthashing"
	pb "github.com/chuamingkai/50.041DynamoProject/pkg/internalcomm"
	"github.com/chuamingkai/50.041DynamoProject/pkg/vectorclock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type GetRepMessage struct {
	peerId    uint64
	repObject models.Object
	err       error
}

type GetRepResult struct {
	Replicas           []models.Object 	`json:"replicas"`
	HasVersionConflict 	bool            `json:"hasVersionConflict"`
	WriteCoordinator	int				`json:"writeCoordinator"`
}

type PutRepMessage struct {
	peerId     uint64
	putSuccess bool
	err        error
}

type PutRepResult struct {
	Success bool  `json:"success"`
	Err     error `json:"error"`
}

func (s *nodesServer) serverGetReplicas(bucketName, key string) (*GetRepResult, error) {
	// Get preference list for the key
	prefList := s.ring.GetPreferenceList(key)

	// Create context to get replicas
	serverDeadline := time.Now().Add(5 * time.Second)
	ctxRep, cancel := context.WithDeadline(context.Background(), serverDeadline)
	defer cancel()

	// Send GetRep request to all servers in preference list
	notifyChan := make(chan GetRepMessage, len(prefList))
	for _, virtualNode := range prefList {
		getRepReq := pb.GetRepRequest{
			BucketName: bucketName,
			Key:        key,
		}

		// Send GetRep request to virtual node
		go func(vn consistenthash.VirtualNode, notifyChan chan GetRepMessage) {
			var getRepRes *pb.GetRepResponse
			var err error

			log.Printf("Getting replica from node %v.\n", vn.NodeId)
			if vn.NodeId == uint64(s.nodeId) {
				getRepRes, err = s.GetReplica(ctxRep, &getRepReq)
			} else {
				peerAddr := fmt.Sprintf("localhost:%v", vn.NodeId)
				dialOptions := grpc.WithTransportCredentials(insecure.NewCredentials())
				var conn *grpc.ClientConn
				conn, err = grpc.Dial(peerAddr, dialOptions)

				if err == nil {
					peer := pb.NewReplicationClient(conn)
					getRepRes, err = peer.GetReplica(ctxRep, &getRepReq)
					if err != nil {
						log.Printf("Error getting replica from node %v: %v\n", vn.NodeId, err.Error())
					}
				} else {
					log.Printf("Error contacting node %v: %v\n", vn.NodeId, err)
				}
				defer conn.Close()
			}

			if err != nil {
				notifyChan <- GetRepMessage{
					peerId: vn.NodeId,
					err:    err,
				}
				return
			}

			var repObject models.Object
			if getRepRes.Data != nil {
				decodeErr := json.Unmarshal(bytes.Trim(getRepRes.Data, "\x00"), &repObject)
				if decodeErr != nil {
					notifyChan <- GetRepMessage{
						peerId: vn.NodeId,
						err:    decodeErr,
					}
				} else {
					notifyChan <- GetRepMessage{
						peerId:    vn.NodeId,
						repObject: repObject,
					}
				}
			} else {
				notifyChan <- GetRepMessage{
					peerId: vn.NodeId,
					err:    fmt.Errorf("failed to get replica from node %v", vn.NodeId),
				}
			}

		}(virtualNode, notifyChan)
	}

	var replicas []models.Object
	writeCoordinator := -1

	successCount := 0
	errorMsg := "ERROR: "
	for i := 0; i < len(prefList); i++ {
		select {
		case notifyMsg := <-notifyChan:
			if notifyMsg.err == nil {
				log.Printf("Received GET replica response from node %v for key {%v}: val {%v}", notifyMsg.peerId, key, notifyMsg.repObject.Value)
				successCount++
				replicas = append(replicas, notifyMsg.repObject)

				// Add write coordinator -> must not be current node!
				if writeCoordinator < 0 && notifyMsg.peerId != uint64(s.internalAddr) {
					writeCoordinator = int(notifyMsg.peerId)
					log.Printf("Write coordinator: %v\n", writeCoordinator)
				}
			} else if e, ok := status.FromError(notifyMsg.err); ok && e.Code() == codes.NotFound {
				log.Printf("Received GET replica response from node %v for key {%v}: key not found", notifyMsg.peerId, key)
				successCount++
			} else {
				errorMsg = errorMsg + notifyMsg.err.Error() + ";"
				log.Printf("Received GET replica response from node %v for key {%v}: Error occurred: %v", notifyMsg.peerId, key, notifyMsg.err.Error())
			}
		case <-ctxRep.Done():
		}
		if ctxRep.Err() != nil || successCount == config.MIN_READS {
			break
		}
	}

	log.Printf("Finished getting replicas for key %v, received %v successful response(s).\n", key, successCount)

	// TODO: Figure out how to handle the case when not enough replicas are received
	if successCount < config.MIN_READS { // Not enough replicas received -> unsuccessful read operation
		return nil, errors.New(errorMsg)
	} else if len(replicas) == 0 { // No replicas received -> key does not exist in database
		return &GetRepResult{Replicas: replicas, HasVersionConflict: false, WriteCoordinator: writeCoordinator}, errors.New("key does not exist")
	} else if allSame(replicas) { // No version conflicts among replicas
		return &GetRepResult{Replicas: replicas, HasVersionConflict: false, WriteCoordinator: writeCoordinator}, nil
	} else { // Versioning conflicts found
		return &GetRepResult{Replicas: replicas, HasVersionConflict: true, WriteCoordinator: writeCoordinator}, nil
	}
}

func (s *nodesServer) serverPutReplicas(bucketName, key string, data []byte) *PutRepResult {
	// Get preference list for the key
	prefList := s.ring.GetPreferenceList(key)

	// Create context to put replicas
	serverDeadline := time.Now().Add(5 * time.Second)
	ctxRep, cancel := context.WithDeadline(context.Background(), serverDeadline)
	defer cancel()

	// Send GetRep request to all servers in preference list
	notifyChan := make(chan PutRepMessage, config.REPLICATION_FACTOR)
	putRepReq := &pb.PutRepRequest{
		Key:        key,
		Data:       data,
		BucketName: bucketName,
		SenderId:   s.nodeId,
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

			if err != nil || putRepRes == nil {
				notifyChan <- PutRepMessage{
					peerId:     vn.NodeId,
					putSuccess: false,
					err:        err,
				}
			} else {
				notifyChan <- PutRepMessage{
					peerId:     vn.NodeId,
					putSuccess: putRepRes.IsDone,
					err:        nil,
				}
			}
		}(virtualNode, notifyChan)
	}

	successCount := 0
	errorMsg := "ERROR MESSAGE: "
	for i := 0; i < len(prefList); i++ {
		select {
		case notifyMsg := <-notifyChan:
			if notifyMsg.err == nil {
				log.Printf("Received PUT replica response from server %v for key {%v} val {%s}: SUCCESS", notifyMsg.peerId, key, data)
				successCount++
			} else {
				errorMsg = errorMsg + notifyMsg.err.Error() + ";"
				log.Printf("Received PUT replica response from server %v for key {%v} val {%s}: ERROR occured: %v\n", notifyMsg.peerId, key, data, notifyMsg.err.Error())
			}
		case <-ctxRep.Done():
		}
		if ctxRep.Err() != nil || successCount == config.REPLICATION_FACTOR {
			break
		}
	}

	log.Printf("Finished PUTTING replica(s) for key {%v}, received {%v} response", key, successCount)
	if successCount == 0 { // All servers failed to save replica
		return &PutRepResult{Success: false, Err: errors.New(errorMsg)}
	} else {
		// TODO: Add case to handle partial success (not all servers in preference list successfuly saved replica)
		return &PutRepResult{Success: true, Err: nil}
	}
}

func allSame(replicas []models.Object) bool {
	clock1 := replicas[0].VC

	for i := 1; i < len(replicas); i++ {
		clock2 := replicas[i].VC
		if vectorclock.IsAncestorOf(clock1, clock2) || vectorclock.IsEqualTo(clock1, clock2) {
			continue
		} else {
			return false
		}
	}
	return true
}
