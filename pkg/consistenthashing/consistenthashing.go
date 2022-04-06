package consistenthash

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"

	config "github.com/chuamingkai/50.041DynamoProject/config"
)

type Ring struct {
	Nodes   *DoublyLinkedList
	NodeMap map[string]NodeInfo
}

type DoublyLinkedList struct {
	Length int
	Head   *VirtualNode
}

type NodeInfo struct {
	NodeId       uint64
	VirtualNodes []*VirtualNode
}

type VirtualNode struct {
	VirtualName string
	NodeId      uint64
	Hash        *big.Int
	Next        *VirtualNode
	Prev        *VirtualNode
}

type ReallocationNotice struct {
	TargetNode *VirtualNode
	NewNode    *VirtualNode
	// the targetNode sends newNode all the keys it has that is >= newNode.Hash
}

// Hash returns the MD5 hash value of the provided string as *big.Int
func Hash(value string) *big.Int {
	data := []byte(value)
	sum := md5.Sum(data)
	bi := big.NewInt(0)
	bi.SetBytes(sum[:])
	return bi
}

func NewRing() *Ring {
	dll := DoublyLinkedList{
		Length: 0,
		Head:   nil,
	}
	nodeMap := make(map[string]NodeInfo)
	return &Ring{&dll, nodeMap}
}

func newVirtualNode(name string, id uint64) *VirtualNode {
	return &VirtualNode{
		VirtualName: name,
		NodeId:      id,
		Hash:        Hash(name),
	}
}

func newNodeInfo(id uint64, ls []*VirtualNode) NodeInfo {
	return NodeInfo{
		NodeId:       id,
		VirtualNodes: ls,
	}
}

// TraverseAndPrint returns a string representing the data stored in the doubly linked list
func (dll DoublyLinkedList) TraverseAndPrint() string {
	var output bytes.Buffer
	output.WriteString(fmt.Sprintf("Linked list of length: %d\n", dll.Length))
	node := dll.Head
	for i := 0; i < dll.Length; i++ {
		output.WriteString(fmt.Sprintf("[%s@%d] %s\n", node.VirtualName, node.NodeId, node.Hash))
		node = node.Next
	}
	return output.String()
}

// AddNode adds a node to the Ring with given name and id
// Caller of AddNode function has to handle reallococation of keys.
// targetNode needs to check for all keys' hash that are >= newNode.Hash
// and send those key-value pairs to newNode
func (r *Ring) AddNode(name string, id uint64) []ReallocationNotice {
	ls := make([]*VirtualNode, config.NUM_VIRTUAL_NODES)
	reAlloc := make([]ReallocationNotice, 0)

	isEmptyBefore := r.Nodes.Length == 0
	for i := 0; i < config.NUM_VIRTUAL_NODES; i++ {
		vName := fmt.Sprintf("%s_%d", name, i)
		vNode := newVirtualNode(vName, id)
		ls[i] = vNode
		r.Nodes.Length++
		if r.Nodes.Length == 1 {
			// linked list was empty
			r.Nodes.Head = vNode
		} else {
			cmpNode := r.Nodes.Head
			for cmpNode.Next != nil {
				if cmpNode.Hash.Cmp(vNode.Hash) == -1 {
					cmpNode = cmpNode.Next
				} else {
					break
				}
			}
			if cmpNode == r.Nodes.Head {
				if vNode.Hash.Cmp(cmpNode.Hash) == -1 {
					// insert as head
					r.Nodes.Head = vNode
					cmpNode.Prev = vNode
					vNode.Next = cmpNode
				} else {
					// insert after head
					vNode.Prev = cmpNode
					vNode.Next = cmpNode.Next
					if cmpNode.Next != nil {
						cmpNode.Next.Prev = vNode
					}
					cmpNode.Next = vNode
				}
			} else {
				if cmpNode.Hash.Cmp(vNode.Hash) != -1 {
					prevNode := cmpNode.Prev
					cmpNode.Prev = vNode
					prevNode.Next = vNode
					vNode.Next = cmpNode
					vNode.Prev = prevNode
				} else {
					// terminated EOL
					cmpNode.Next = vNode
					vNode.Prev = cmpNode
				}
			}
		}
	}
	r.NodeMap[name] = newNodeInfo(id, ls)

	// reallocation
	if !isEmptyBefore {
		for _, vNode := range ls {
			var target *VirtualNode
			if vNode.Prev == nil {
				// prev is tail
				target = vNode.Next
				for target.Next != nil {
					target = target.Next
				}
			} else {
				target = vNode.Prev
			}
			if target.NodeId == vNode.NodeId {
				continue
			}
			reAlloc = append(reAlloc, ReallocationNotice{
				TargetNode: target,
				NewNode:    vNode,
			})
		}
	}

	return reAlloc
}

// SearchKey returns the first (coordinator) Node that is expected to store that key
func (r *Ring) SearchKey(key string) VirtualNode {
	hashedKey := Hash(key)
	node := r.Nodes.Head
	for node.Next != nil {
		if node.Hash.Cmp(hashedKey) == -1 {
			node = node.Next
		} else {
			break
		}
	}
	if node.Prev == nil {
		// return tail
		for node.Next != nil {
			node = node.Next
		}
		return *node
	}
	return *node.Prev
}

// GetPreferenceList returns the preference list for key
func (r *Ring) GetPreferenceList(key string) []VirtualNode {
	prefList := make([]VirtualNode, 0)
	node := r.SearchKey(key)
	prefList = append(prefList, node)
	num_replicate_possible := config.REPLICATION_FACTOR
	if len(r.NodeMap) < config.REPLICATION_FACTOR {
		num_replicate_possible = len(r.NodeMap)
	}
	for i := 0; i < num_replicate_possible-1; i++ {
		fmt.Println(node)
		if node.Next == nil {
			node = *r.Nodes.Head
		} else {
			node = *node.Next
		}
		hit := false
		for _, n := range prefList {
			if node.NodeId == n.NodeId {
				hit = true
				break
			}
		}
		if hit {
			i--
			continue
		} else {
			prefList = append(prefList, node)
		}
	}
	return prefList
}

// IsNodeResponsibleForKey returns true if the physical node is in
// the preference list for that key
func (r *Ring) IsNodeResponsibleForKey(key string, id uint64) bool {
	prefList := r.GetPreferenceList(key)
	for _, n := range prefList {
		fmt.Println(n)
		if n.NodeId == id {
			return true
		}
	}
	return false
}

// RemoveNode removes the node with the provided name
// Reallocation of keys needs to be handled by caller
func (r *Ring) RemoveNode(name string) {
	node, prs := r.NodeMap[name]
	if !prs {
		return
	}
	for _, vNode := range node.VirtualNodes {
		if vNode.Prev != nil {
			vNode.Prev.Next = vNode.Next
		} else if vNode == r.Nodes.Head {
			r.Nodes.Head = vNode.Next
		}
		if vNode.Next != nil {
			vNode.Next.Prev = vNode.Prev
		}
		r.Nodes.Length--
	}
	delete(r.NodeMap, name)
}

// ImportRingFromFile constructs a Ring based on data from a file specified by filename
// File Format:
// name,id\n
// name,id\n
// name,id\n
// ...
// \n
// vname,id\n
// vname,id\n
// ...
func ImportRingFromFile(filename string) (*Ring, error) {
	r := NewRing()
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	tmpMap := make(map[uint64]string)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			break
		}
		split := strings.Split(line, ",")
		id, e := strconv.ParseUint(split[1], 10, 64)
		if e != nil {
			return nil, e
		}
		tmpMap[id] = split[0]
		r.NodeMap[split[0]] = newNodeInfo(id, make([]*VirtualNode, 0))
	}

	var prev *VirtualNode
	for scanner.Scan() {
		line := scanner.Text()
		if line == "\n" {
			break
		}
		split := strings.Split(line, ",")
		id, e := strconv.ParseUint(split[1], 10, 64)
		if e != nil {
			return nil, e
		}
		curr := newVirtualNode(split[0], id)
		if prev != nil {
			prev.Next = curr
			curr.Prev = prev
		}
		if r.Nodes.Head == nil {
			r.Nodes.Head = curr
		}
		physical := r.NodeMap[tmpMap[id]]
		physical.VirtualNodes = append(physical.VirtualNodes, curr)
		prev = curr
		r.Nodes.Length++
	}
	return r, nil
}

// BackupRing outputs the current ring state into a file specified by filename
// File Format:
// name,id\n
// name,id\n
// name,id\n
// ...
// \n
// vname,id\n
// vname,id\n
// ...
func (r *Ring) BackupRing(filename string) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0744)
	if err != nil {
		return err
	}
	defer f.Close()

	for k, v := range r.NodeMap {
		f.WriteString(fmt.Sprintf("%s,%d\n", k, v.NodeId))
	}

	f.WriteString("\n")
	node := r.Nodes.Head
	for node.Next != nil {
		f.WriteString(fmt.Sprintf("%s,%d\n", node.VirtualName, node.NodeId))
		node = node.Next
	}
	f.WriteString(fmt.Sprintf("%s,%d\n", node.VirtualName, node.NodeId))
	return nil
}
