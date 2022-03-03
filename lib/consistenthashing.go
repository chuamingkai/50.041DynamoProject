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

// hash returns the MD5 hash value of the provided value as *big.Int
func hash(value string) *big.Int {
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
		Hash:        hash(name),
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
func (r *Ring) AddNode(name string, IP string) {
	ls := make([]*VirtualNode, NUM_VIRTUAL_NODES)

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
			// TODO: redistribution
			/*
				inform prevNode of node's addition, tell prevNode node.Hash
				prevNodes needs to check for all keys' hash that are >= node.Hash
				and send those key-value pairs to node
			*/
		}
	}
	r.NodeMap[name] = newNodeInfo(IP, ls)
}

// SearchKey returns the Node that is expected to store that key
func (r *Ring) SearchKey(key string) VirtualNode {
	hashedKey := hash(key)
	node := r.Nodes.Head
	for {
		if node.Hash.Cmp(hashedKey) == -1 {
			node = node.Next
		} else {
			return *node.Prev
		}
	}
}

// RemoveNode removes the node with the provided name
func (r *Ring) RemoveNode(name string) {
	// TODO incorporate virtual nodes
	if r.Nodes.Length == 0 {
		return
	}
	hashedName := hash(name)
	if r.Nodes.Head.Hash.Cmp(hashedName) == 0 {
		if r.Nodes.Length == 1 {
			r.Nodes.Head = nil
		} else {
			r.Nodes.Head.Next.Prev = r.Nodes.Head.Prev
			r.Nodes.Head.Prev.Next = r.Nodes.Head.Next
			r.Nodes.Head = r.Nodes.Head.Next
		}
		r.Nodes.Length--
		return
	}
	if r.Nodes.Length > 1 {
		cmpNode := r.Nodes.Head.Next
		for i := 0; i < r.Nodes.Length-1; i++ {
			if cmpNode.Hash.Cmp(hashedName) == 0 {
				cmpNode.Prev.Next = cmpNode.Next
				cmpNode.Next.Prev = cmpNode.Prev
				r.Nodes.Length--
				break
			}
			cmpNode = cmpNode.Next
		}
	}
}
