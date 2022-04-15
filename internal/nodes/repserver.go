package nodes

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
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
		conflictingReps := getConflictingReps(replicas)
		return &GetRepResult{
			Replicas:           conflictingReps,
			HasVersionConflict: len(conflictingReps) > 1,
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
				s.initiateHintedHandoff(vn, bucketName, data)
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

// Get list of conflicting replicas
func getConflictingReps(replicas []models.Object) []models.Object {
	output := make([]models.Object, 0, len(replicas))

	output = append(output, replicas[0])
	targetVClock := replicas[0].VC

	for i := 1; i < len(replicas); i++ {
		replicaVClock := replicas[i].VC
		if vectorclock.IsEqualTo(replicaVClock, targetVClock) || vectorclock.IsDescendantOf(targetVClock, replicaVClock) {
			continue
		} else if vectorclock.IsDescendantOf(replicaVClock, targetVClock) {
			targetVClock = replicaVClock // Drop current target and replace with descendant

			// Remove remaining ancestors
			for j := 0; j < len(output); j++ {
				if vectorclock.IsAncestorOf(output[j].VC, targetVClock) {
					output = append(output[:j], output[:j+1]...)
				}
			}
		} else if vectorclock.IsConcurrentWith(replicaVClock, targetVClock) {
			output = append(output, replicas[i])
		}
	}

	return output
}

func (s *nodesServer) initiateHintedHandoff(originalTarget consistenthash.VirtualNode, bucketName string, replica []byte) {
	i := strings.LastIndex(originalTarget.VirtualName, "_")
	origNode := originalTarget.VirtualName[:i]
	var o models.Object
	err := json.Unmarshal(replica, &o)
	if err != nil {
		log.Printf("HintedHandoff: Error unmarshalling object for hinted handoff")
		return
	}
	hint := models.HintedObject{Data: o, BucketName: bucketName}
	nextNode := originalTarget.Next
	for nextNode != nil {
		if nextNode.NodeId == uint64(s.internalAddr) || nextNode.NodeId == originalTarget.NodeId {
			nextNode = nextNode.Next
			continue
		}
		isSuccess := s.sendHint(origNode, nextNode.NodeId, hint)
		if isSuccess {
			return
		}
		nextNode = nextNode.Next
	}
	nextNode = s.ring.Nodes.Head
	for nextNode.VirtualName != originalTarget.VirtualName {
		if nextNode.VirtualName == originalTarget.VirtualName {
			return
		}
		if nextNode.NodeId == uint64(s.internalAddr) || nextNode.NodeId == originalTarget.NodeId {
			nextNode = nextNode.Next
			continue
		}
		isSuccess := s.sendHint(origNode, nextNode.NodeId, hint)
		if isSuccess {
			return
		}
		nextNode = nextNode.Next
	}
}

func (s *nodesServer) sendHint(targetNode string, receiverAddr uint64, hint models.HintedObject) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("HintedHandoff: Receiver of hint: %d\n", receiverAddr)
	peerAddr := fmt.Sprintf("localhost:%v", receiverAddr)
	dialOptions := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.DialContext(ctx, peerAddr, dialOptions)
	if err != nil {
		log.Printf("HintedHandoff: Error contacting node %v: %v\n", receiverAddr, err)
		return false
	}
	defer conn.Close()

	peer := pb.NewReplicationClient(conn)
	hint_bytes, err := json.Marshal(hint)
	if err != nil {
		log.Println("HintedHandoff: Error marshalling hinted object")
		return false
	}
	hintResponse, err := peer.HintedHandoff(ctx, &pb.HintHandoffRequest{TargetNode: targetNode, Data: hint_bytes})
	if err != nil {
		log.Printf("HintedHandoff: Error sending hint to node %v: %v\n", receiverAddr, err.Error())
		return false
	}
	return hintResponse.IsDone
}

/*addnode and realloc keys to the respective nodes*/
func (s *nodesServer) serverReallocKeys(nodename, bucketName string, portno uint64) bool {
	var nodes []uint64
	notice := s.ring.AddNode(nodename, portno)
	log.Printf("Node %v added to ring.\n", nodename)
	log.Println(s.ring.Nodes.TraverseAndPrint())
	log.Println(notice)
	if len(notice) > 0 {
		//delete := make([]models.HintedObject, 0)
		for _, node := range notice {
			//fmt.Println("TargetNode", *node.TargetNode)
			//fmt.Println("TargetNode", *node.TargetNode)
			done := false
			//log.Println(node.TargetNode.NodeId, uint64(s.nodeId))
			if node.TargetNode.NodeId == uint64(s.nodeId)-3000 {
				for _, v := range nodes {
					if v == node.NewNode.NodeId {
						done = true
						break
					}
				}

				if !done {

					hashnode := node.NewNode
					for i := 0; i < config.REPLICATION_FACTOR-1; i++ {
						hashnode = hashnode.Prev
					}
					bucketNames, errb := s.boltDB.GetAllBuckets()
					if errb == nil {
						for _, bucketname := range bucketNames {
							if bucketname == "hints" {
								continue
							}
							log.Printf("reallocating keys to %v from %s bucket \n", node.NewNode.NodeId, bucketname)
							err := s.boltDB.Iterate(bucketname, func(k, v []byte) error {

								//destinationNodename := string(k[:])
								targetPort := node.NewNode.NodeId
								//fmt.Println(string(k[:]))
								//fmt.Printf("TargetPort %v NodeHash %v KeyHash%v\n", targetPort, node.NewNode.Hash, consistenthash.Hash(string(k[:])))
								var hintedDatas []models.HintedObject
								var reallocObj models.Object
								/*transfer keys including responsible replicas*/
								if hashnode.Hash.Cmp(consistenthash.Hash(string(k[:]))) <= 0 {

									if err := json.Unmarshal(v, &reallocObj); err != nil {
										//fmt.Println(reallocObj)

										return err
									} else {
										//fmt.Println(reallocObj)
										hintedDatas = append(hintedDatas, models.HintedObject{BucketName: bucketname, Data: reallocObj})
										//fmt.Println(hintedDatas)
									}
									/*delete keys that are now responsible by new node*/
									//if node.NewNode.Hash.Cmp(consistenthash.Hash(string(k[:]))) <= 0 {
									//	delete = append(delete, models.HintedObject{BucketName: bucketname, Data: reallocObj})

									//}

								}

								if s.clientPutMultiple(targetPort, hintedDatas) {
									//delete = append(delete, reallocObj.Key)

								} else {
									log.Printf("Reallocation error")
								}
								return nil

							})
							if err != nil {
								log.Printf("Reallocation error: %s\n", err)

							}
							nodes = append(nodes, node.NewNode.NodeId)
						}
					}
				}
			}

		}
		//fmt.Println("delete", delete)
		/*
			for _, v := range delete {
				err := s.boltDB.DeleteKey(v.BucketName, v.Data.Key)
				if v.Data.Key != "" {
					log.Printf("Successfully handed off key %s from bucket %s\n", v.Data.Key, v.BucketName)
				}
				if err != nil {
					log.Printf("Reallocation error: %s\n", err)
				}
			}
		*/
	}
	return true

}

func (s *nodesServer) delnodeReallocKeys(nodename string) {
	s.ring.RemoveNode(nodename)
	for _, n := range s.ring.NodeMap {
		hashnode := n.VirtualNodes[0]
		for i := 0; i < config.REPLICATION_FACTOR-1; i++ {
			hashnode = hashnode.Prev
		}
		bucketNames, errb := s.boltDB.GetAllBuckets()
		if errb == nil {
			for _, bucketname := range bucketNames {
				if bucketname == "hints" {
					continue
				}
				log.Printf("reallocating keys from %s bucket to %v\n", bucketname, n.NodeId)
				err := s.boltDB.Iterate(bucketname, func(k, v []byte) error {

					//destinationNodename := string(k[:])
					targetPort := n.NodeId
					//fmt.Println(string(k[:]))
					//fmt.Printf("TargetPort %v NodeHash %v KeyHash%v\n", targetPort, node.NewNode.Hash, consistenthash.Hash(string(k[:])))
					var hintedDatas []models.HintedObject
					var reallocObj models.Object
					/*transfer keys including responsible replicas*/
					if hashnode.Hash.Cmp(consistenthash.Hash(string(k[:]))) <= 0 {

						if err := json.Unmarshal(v, &reallocObj); err != nil {
							//fmt.Println(reallocObj)

							return err
						} else {
							//fmt.Println(reallocObj)
							hintedDatas = append(hintedDatas, models.HintedObject{BucketName: bucketname, Data: reallocObj})
							//fmt.Println(hintedDatas)
						}
						/*delete keys that are now responsible by new node*/
						//if node.NewNode.Hash.Cmp(consistenthash.Hash(string(k[:]))) <= 0 {
						//	delete = append(delete, models.HintedObject{BucketName: bucketname, Data: reallocObj})

						//}

					}

					if s.clientPutMultiple(targetPort, hintedDatas) {
						//delete = append(delete, reallocObj.Key)

					} else {
						log.Printf("Reallocation error")
					}
					return nil

				})
				if err != nil {
					log.Printf("Reallocation error: %s\n", err)

				}
			}
		}

	}

}
