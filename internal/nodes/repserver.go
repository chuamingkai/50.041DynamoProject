package nodes

import (
	"bytes"
	"context"
	"crypto/rand"
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
	"google.golang.org/grpc/credentials/insecure"
)

type SuccessStatus int

const (
	UNSUCCESSFUL SuccessStatus = iota
	PARTIAL_SUCCESS
	FULL_SUCCESS
)

type GetRepMessage struct {
	peerId   uint64
	repBytes []byte
	err      error
}

type GetRepResult struct {
	Replicas           []models.Object `json:"replicas"`
	HasVersionConflict bool            `json:"hasVersionConflict"`
	WriteCoordinator   int             `json:"writeCoordinator"`
	SuccessStatus      SuccessStatus   `json:"successStatus"`
}

type PutRepMessage struct {
	peerId     uint64
	putSuccess bool
	err        error
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
				if err != nil {
					log.Printf("Error contacting node %v: %v\n", vn.NodeId, err)
				} else {
					defer conn.Close()
					peer := pb.NewReplicationClient(conn)
					getRepRes, err = peer.GetReplica(ctxRep, &getRepReq)
					if err != nil {
						log.Printf("Error getting replica from node %v: %v\n", vn.NodeId, err.Error())
					}
				}

			}

			if err != nil {
				notifyChan <- GetRepMessage{
					peerId:   vn.NodeId,
					err:      err,
					repBytes: nil,
				}
			} else {
				notifyChan <- GetRepMessage{
					peerId:   vn.NodeId,
					err:      nil,
					repBytes: getRepRes.Data,
				}
			}

		}(virtualNode, notifyChan)
	}

	var replicas []models.Object
	writeCoordinator := -1

	responseCount := 0
	for i := 0; i < len(prefList); i++ {
		select {
		case notifyMsg := <-notifyChan:
			if notifyMsg.err != nil {
				log.Printf("Received GET replica response from node %v for key {%v}: Error occurred: %v", notifyMsg.peerId, key, notifyMsg.err.Error())
			} else if notifyMsg.repBytes != nil {
				responseCount++
				var repObject models.Object
				json.Unmarshal(bytes.Trim(notifyMsg.repBytes, "\x00"), &repObject)
				log.Printf("Received GET replica response from node %v for key {%v}: val {%v}", notifyMsg.peerId, key, repObject.Value)
				replicas = append(replicas, repObject)

				if writeCoordinator < 0 && notifyMsg.peerId != uint64(s.nodeId) {
					writeCoordinator = int(notifyMsg.peerId) + 3000
					log.Printf("Write coordinator for key {%v}: %v\n", key, writeCoordinator)
				}
			} else {
				log.Printf("Received GET replica response from node %v for key {%v}: key not found", notifyMsg.peerId, key)
				responseCount++
			}
		case <-ctxRep.Done():
		}
		if ctxRep.Err() != nil || responseCount == config.MIN_READS {
			break
		}
	}

	log.Printf("Finished getting replicas for key %v, received %v successful response(s).\n", key, responseCount)

	if responseCount == 0 { // All servers in preference list are dead
		return &GetRepResult{
			SuccessStatus: UNSUCCESSFUL,
		}, errors.New("all servers in preference list are down")
	} else if responseCount < config.MIN_READS { // Not enough replicas received -> partially successful read operation
		return &GetRepResult{
			SuccessStatus: UNSUCCESSFUL,
		}, errors.New("all servers in preference list are down")
	} else if len(replicas) == 0 { // No replicas received -> key does not exist in database
		return &GetRepResult{
			Replicas:      replicas,
			SuccessStatus: UNSUCCESSFUL,
		}, nil
	} else {
		return &GetRepResult{
			Replicas:           replicas,
			HasVersionConflict: haveVersionConflicts(replicas),
			WriteCoordinator:   writeCoordinator,
			SuccessStatus:      FULL_SUCCESS,
		}, nil
	}
}

func (s *nodesServer) serverPutReplicas(bucketName, key string, data []byte) (SuccessStatus, error) {
	// Get preference list for the key
	prefList := s.ring.GetPreferenceList(key)

	// Create context to put replicas
	serverDeadline := time.Now().Add(5 * time.Second)
	ctxRep, cancel := context.WithDeadline(context.Background(), serverDeadline)
	defer cancel()

	// Send PutRep request to all servers in preference list
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
				var conn *grpc.ClientConn
				conn, err = grpc.Dial(peerAddr, dialOptions)
				if err != nil {
					log.Printf("Error contacting node %v: %v\n", vn.NodeId, err)
				} else {
					defer conn.Close()
					peer := pb.NewReplicationClient(conn)
					putRepRes, err = peer.PutReplica(ctxRep, putRepReq)
				}
			}

			if err != nil {
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
	for i := 0; i < len(prefList); i++ {
		select {
		case notifyMsg := <-notifyChan:
			if notifyMsg.err == nil {
				log.Printf("Received PUT replica response from server %v for key {%v} val {%s}: SUCCESS", notifyMsg.peerId, key, data)
				successCount++
			} else {
				log.Printf("Received PUT replica response from server %v for key {%v} val {%s}: ERROR occured: %v\n", notifyMsg.peerId, key, data, notifyMsg.err.Error())
			}
		case <-ctxRep.Done():
		}
		if ctxRep.Err() != nil || successCount == config.MIN_WRITES {
			break
		}
	}

	log.Printf("Finished PUTTING replica(s) for key {%v}, received {%v} response", key, successCount)

	if successCount == 0 { // All servers in preference list are down
		return UNSUCCESSFUL, errors.New("all servers in preference list are down")
	} else if successCount < config.MIN_WRITES { // Partial success; only some servers in preference list successfuly saved replica
		return PARTIAL_SUCCESS, errors.New("not enough replicas saved")
	} else {
		return FULL_SUCCESS, nil
	}
}

func (s *nodesServer) clientPutMultiple(target uint64, datas []models.HintedObject) bool {
	serverDeadline := time.Now().Add(10 * time.Second)
	ctxRep, cancel := context.WithDeadline(context.Background(), serverDeadline)
	defer cancel()

	peerAddr := fmt.Sprintf("localhost:%v", target)
	dialOptions := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.Dial(peerAddr, dialOptions)
	if err != nil {
		log.Printf("Error contacting node %v: %v\n", target+3000, err)
	}
	defer conn.Close()
	client := pb.NewReplicationClient(conn)
	stream, err := client.PutMultiple(ctxRep)
	if err != nil {
		log.Println("Error beginning client streaming: ", err)
	}
	for _, d := range datas {
		bytes_data, err := json.Marshal(d.Data)
		if err != nil {
			log.Println("Error marshalling data")
		}
		err = stream.Send(&pb.MultiPutRequest{Key: d.Data.Key, Data: bytes_data, BucketName: d.BucketName, SenderId: s.nodeId})
		if err != nil {
			log.Println("Error sending data to stream: ", err)
		}
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Println("Error encountered on CloseAndRecv: ", err)
	}
	return reply.IsDone
}

func (s *nodesServer) sendHeartbeat(target uint64) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	peerAddr := fmt.Sprintf("localhost:%v", target)
	dialOptions := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.DialContext(ctx, peerAddr, dialOptions)
	if err != nil {
		log.Printf("Error contacting node %v: %v\n", target+3000, err)
		return false
	}
	log.Println("Node contacted")
	defer conn.Close()

	peer := pb.NewReplicationClient(conn)
	token := make([]byte, 5)
	_, err = rand.Read(token)
	if err != nil {
		log.Println("Error generating heartbeat message")
		return false
	}
	log.Println("heartbeat message generated")
	hrtResponse, err := peer.Heartbeat(ctx, &pb.HeartbeatRequest{Data: token})
	if err != nil {
		log.Printf("Error heartbeat in node %v: %v\n", target+3000, err.Error())
		return false
	}
	return bytes.Equal(hrtResponse.Data, token)

}

// Check for version conflicts among replicas received in GET operation
func haveVersionConflicts(replicas []models.Object) bool {
	result := false
	clock1 := replicas[0].VC

	for i := 1; i < len(replicas); i++ {
		clock2 := replicas[i].VC
		if !vectorclock.IsConcurrentWith(clock1, clock2) {
			result = true
		}
	}
	return result
}

func (s *nodesServer) putInHintBucket(origNode string, origBucket string, hint models.Object) error {
	data, err := s.boltDB.Get(config.HINT_BUCKETNAME, origNode)
	if err != nil {
		return err
	}
	var hintedDatas []models.HintedObject
	if data == nil {
		hintedDatas = []models.HintedObject{{Data: hint, BucketName: origBucket}}
	} else {
		if err := json.Unmarshal(data, &hintedDatas); err != nil {
			return err
		}
		hintedDatas = append(hintedDatas, models.HintedObject{Data: hint, BucketName: origBucket})
	}
	b, err := json.Marshal(hintedDatas)
	if err != nil {
		return err
	}
	if err = s.boltDB.Put(config.HINT_BUCKETNAME, origNode, b); err != nil {
		return err
	}
	return nil
}
