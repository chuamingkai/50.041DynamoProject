package consistenthash

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"math/big"
)

const NUM_VIRTUAL_NODES int = 3 // TODO: probably slap this in some config file later

type Ring struct {
	Nodes   *DoublyLinkedList
	NodeMap map[string]NodeInfo
}

type DoublyLinkedList struct {
	Length int
	Head   *VirtualNode
}

type NodeInfo struct {
	NodeIP       string
	VirtualNodes []*VirtualNode
}

type VirtualNode struct {
	VirtualName string
	NodeIP      string
	Hash        *big.Int
	Next        *VirtualNode
	Prev        *VirtualNode
}

type ReallocationNotice struct {
	targetNode *VirtualNode
	newNode    *VirtualNode
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

func newVirtualNode(name string, IP string) *VirtualNode {
	return &VirtualNode{
		VirtualName: name,
		NodeIP:      IP,
		Hash:        Hash(name),
	}
}

func newNodeInfo(IP string, ls []*VirtualNode) NodeInfo {
	return NodeInfo{
		NodeIP:       IP,
		VirtualNodes: ls,
	}
}

// TraverseAndPrint returns a string representing the data stored in the doubly linked list
func (dll DoublyLinkedList) TraverseAndPrint() string {
	var output bytes.Buffer
	output.WriteString(fmt.Sprintf("Linked list of length: %d\n", dll.Length))
	node := dll.Head
	for i := 0; i < dll.Length; i++ {
		output.WriteString(fmt.Sprintf("[%s@%s] %s\n", node.VirtualName, node.NodeIP, node.Hash))
		node = node.Next
	}
	return output.String()
}

// AddNode adds a node to the Ring with given name and IP
// Caller of AddNode function has to handle reallococation of keys.
// targetNode needs to check for all keys' hash that are >= newNode.Hash
// and send those key-value pairs to newNode
func (r *Ring) AddNode(name string, IP string) []ReallocationNotice {
	ls := make([]*VirtualNode, NUM_VIRTUAL_NODES)
	reAlloc := make([]ReallocationNotice, 0)

	isEmptyBefore := r.Nodes.Length == 0
	for i := 0; i < NUM_VIRTUAL_NODES; i++ {
		vName := fmt.Sprintf("%s_%d", name, i)
		vNode := newVirtualNode(vName, IP)
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
					cmpNode.Next.Prev = vNode
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

			// reallocation
			if !isEmptyBefore {
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
				if target.NodeIP == vNode.NodeIP {
					continue
				}
				reAlloc = append(reAlloc, ReallocationNotice{
					targetNode: target,
					newNode:    vNode,
				})
			}
		}
	}
	r.NodeMap[name] = newNodeInfo(IP, ls)
	return reAlloc
}

// SearchKey returns the first Node that is expected to store that key
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