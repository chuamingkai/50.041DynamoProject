package consistenthash

import (
	"fmt"
	"math/big"
	"sort"
	"testing"
)

type debugNode struct {
	id   string
	hash *big.Int
}

type debugRing []debugNode

func testRingIntegrity(r *Ring) bool {
	forward := make([]string, r.Nodes.Length)
	backward := make([]string, r.Nodes.Length)

	i := 0
	curr := r.Nodes.Head
	for curr.Next != nil {
		forward[i] = curr.VirtualName
		curr = curr.Next
		i++
	}
	forward[i] = curr.VirtualName

	for curr.Prev != nil {
		backward[i] = curr.VirtualName
		curr = curr.Prev
		i--
	}
	backward[i] = curr.VirtualName

	for i := 0; i < len(forward); i++ {
		if forward[i] != backward[i] {
			return false
		}
	}
	return true
}

func TestSingleAdd(t *testing.T) {
	name := "A"
	a := NewRing()
	a.AddNode(name, "10.0.9.1")
	if !testRingIntegrity(a) {
		t.Error("Failed ring integrity check")
	}

	dRing := debugRing{}
	for i := 0; i < NUM_VIRTUAL_NODES; i++ {
		nodeName := fmt.Sprintf("%s_%d", name, i)
		dRing = append(dRing, debugNode{
			id:   nodeName,
			hash: hash(nodeName),
		})
	}

	sort.Slice(dRing, func(i, j int) bool {
		return dRing[i].hash.Cmp(dRing[j].hash) == -1
	})

	if a.Nodes.Length != NUM_VIRTUAL_NODES {
		t.Error("Number of nodes not matching")
	}

	curr := a.Nodes.Head
	for i := 0; i < a.Nodes.Length; i++ {
		if curr.Hash.Cmp(dRing[i].hash) != 0 {
			dRingOrder := ""
			for _, n := range dRing {
				dRingOrder += fmt.Sprintf("[%s] %d\n", n.id, n.hash)
			}

			t.Errorf("Ordering error in ring: \nExpected order:\n%vGiven order: \n%s \n", dRingOrder, a.Nodes.TraverseAndPrint())
			break
		}
		curr = curr.Next
	}
}

func TestMultipleAdd(t *testing.T) {
	names := []string{"A", "B"}
	ip := []string{"10.0.9.1", "10.0.9.2"}
	a := NewRing()
	dRing := debugRing{}

	for i, name := range names {
		a.AddNode(name, ip[i])
		for j := 0; j < NUM_VIRTUAL_NODES; j++ {
			nodeName := fmt.Sprintf("%s_%d", name, j)
			dRing = append(dRing, debugNode{
				id:   nodeName,
				hash: hash(nodeName),
			})
		}
	}
	if !testRingIntegrity(a) {
		t.Error("Failed ring integrity check")
	}

	sort.Slice(dRing, func(i, j int) bool {
		return dRing[i].hash.Cmp(dRing[j].hash) == -1
	})

	if a.Nodes.Length != len(names)*NUM_VIRTUAL_NODES {
		t.Error("Number of nodes not matching")
	}

	curr := a.Nodes.Head
	for i := 0; i < a.Nodes.Length; i++ {
		if curr.Hash.Cmp(dRing[i].hash) != 0 {
			dRingOrder := ""
			for _, n := range dRing {
				dRingOrder += fmt.Sprintf("[%s] %d\n", n.id, n.hash)
			}

			t.Errorf("Ordering error in ring: \nExpected order:\n%vGiven order: \n%s \n", dRingOrder, a.Nodes.TraverseAndPrint())
			break
		}
		curr = curr.Next
	}
}
