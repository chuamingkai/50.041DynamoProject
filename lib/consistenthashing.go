package consistenthash

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"math/big"
)

type Ring struct {
	Nodes *CircularDoublyLinkedList
}

type CircularDoublyLinkedList struct {
	Length int
	Head   *Node
}

type Node struct {
	NodeName string
	NodeIP   string
	Hash     *big.Int
	Next     *Node
	Prev     *Node
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
	dll := CircularDoublyLinkedList{
		Length: 0,
		Head:   nil,
	}
	return &Ring{&dll}
}

func newNode(name string, IP string) *Node {
	return &Node{
		NodeName: name,
		NodeIP:   IP,
		Hash:     hash(name),
	}
}

// TraverseAndPrint returns a string representing the data stored in the doubly linked list
func (dll CircularDoublyLinkedList) TraverseAndPrint() string {
	var output bytes.Buffer
	output.WriteString(fmt.Sprintf("Linked list of length: %d\n", dll.Length))
	node := dll.Head
	for i := 0; i < dll.Length; i++ {
		output.WriteString(fmt.Sprintf("[%s@%s] %s\n", node.NodeName, node.NodeIP, node.Hash))
	}
	return output.String()
}

func (r *Ring) AddNode(name string, IP string) {
	// TODO virtual nodes
	node := newNode(name, IP)
	r.Nodes.Length++
	if r.Nodes.Length == 1 {
		node.Next = node
		node.Prev = node
		r.Nodes.Head = node
	} else {
		cmpNode := r.Nodes.Head
		for i := 0; i < r.Nodes.Length; i++ {
			if cmpNode.Hash.Cmp(node.Hash) == -1 {
				cmpNode = cmpNode.Next
			} else {
				break
			}
		}
		prevNode := cmpNode.Prev
		cmpNode.Prev = node
		prevNode.Next = node
		node.Next = cmpNode
		node.Prev = prevNode
		// TODO: redistribution
		/*
			inform prevNode of node's addition, tell prevNode node.Hash
			prevNodes needs to check for all keys' hash that are >= node.Hash
			and send those key-value pairs to node
		*/
	}
}

// SearchKey returns the first Node that is expected to store that key
func (r *Ring) SearchKey(key string) Node {
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

func (r *Ring) RemoveNode(name string) {
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
