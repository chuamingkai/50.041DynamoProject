package consistenthash

import (
	"fmt"
	"math/big"
	"sort"
	"testing"

	config "github.com/chuamingkai/50.041DynamoProject/config"
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
	if err := config.LoadEnvFile(); err != nil {
		t.Errorf("Error loading .env file: %v\n", err.Error())
	}
	name := "A"
	a := NewRing()
	a.AddNode(name, 8000)
	if !testRingIntegrity(a) {
		t.Error("Failed ring integrity check")
	}

	dRing := debugRing{}
	for i := 0; i < config.NUM_VIRTUAL_NODES; i++ {
		nodeName := fmt.Sprintf("%s_%d", name, i)
		dRing = append(dRing, debugNode{
			id:   nodeName,
			hash: Hash(nodeName),
		})
	}

	sort.Slice(dRing, func(i, j int) bool {
		return dRing[i].hash.Cmp(dRing[j].hash) == -1
	})

	if a.Nodes.Length != config.NUM_VIRTUAL_NODES {
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
	if err := config.LoadEnvFile(); err != nil {
		t.Errorf("Error loading .env file: %v\n", err.Error())
	}
	names := []string{"A", "B", "C", "D"}
	ip := []uint64{8000, 8001, 8002, 8003}
	a := NewRing()
	dRing := debugRing{}

	for i, name := range names {
		a.AddNode(name, ip[i])
		for j := 0; j < config.NUM_VIRTUAL_NODES; j++ {
			nodeName := fmt.Sprintf("%s_%d", name, j)
			dRing = append(dRing, debugNode{
				id:   nodeName,
				hash: Hash(nodeName),
			})
		}
	}
	if !testRingIntegrity(a) {
		t.Error("Failed ring integrity check")
	}

	sort.Slice(dRing, func(i, j int) bool {
		return dRing[i].hash.Cmp(dRing[j].hash) == -1
	})

	if a.Nodes.Length != len(names)*config.NUM_VIRTUAL_NODES {
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

func TestSingleAddAndDelete(t *testing.T) {
	if err := config.LoadEnvFile(); err != nil {
		t.Errorf("Error loading .env file: %v\n", err.Error())
	}
	name := "A"
	a := NewRing()
	a.AddNode(name, 8000)
	if !testRingIntegrity(a) {
		t.Error("Failed ring integrity check")
	}

	dRing := debugRing{}
	for i := 0; i < config.NUM_VIRTUAL_NODES; i++ {
		nodeName := fmt.Sprintf("%s_%d", name, i)
		dRing = append(dRing, debugNode{
			id:   nodeName,
			hash: Hash(nodeName),
		})
	}

	sort.Slice(dRing, func(i, j int) bool {
		return dRing[i].hash.Cmp(dRing[j].hash) == -1
	})

	if a.Nodes.Length != config.NUM_VIRTUAL_NODES {
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

	a.RemoveNode(name)
	if a.Nodes.Length != 0 {
		t.Error("Deleting only node does not correctly reduce length")
	}
	if a.Nodes.Head != nil {
		t.Error("Deleting only node does not correctly clear head of linked list")
	}
	if _, ok := a.NodeMap[name]; ok {
		t.Error("Deleting node fails to clear info from NodeMap")
	}
}

func TestMultipleAddAndDelete(t *testing.T) {
	if err := config.LoadEnvFile(); err != nil {
		t.Errorf("Error loading .env file: %v\n", err.Error())
	}
	names := []string{"A", "B"}
	ip := []uint64{8000, 8001}
	a := NewRing()
	dRing := debugRing{}

	for i, name := range names {
		a.AddNode(name, ip[i])
		for j := 0; j < config.NUM_VIRTUAL_NODES; j++ {
			nodeName := fmt.Sprintf("%s_%d", name, j)
			dRing = append(dRing, debugNode{
				id:   nodeName,
				hash: Hash(nodeName),
			})
		}
	}
	if !testRingIntegrity(a) {
		t.Error("Failed ring integrity check")
	}

	sort.Slice(dRing, func(i, j int) bool {
		return dRing[i].hash.Cmp(dRing[j].hash) == -1
	})

	if a.Nodes.Length != len(names)*config.NUM_VIRTUAL_NODES {
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

	a.RemoveNode(names[0])
	dRing = debugRing{}
	for j := 0; j < config.NUM_VIRTUAL_NODES; j++ {
		nodeName := fmt.Sprintf("%s_%d", names[1], j)
		dRing = append(dRing, debugNode{
			id:   nodeName,
			hash: Hash(nodeName),
		})
	}
	sort.Slice(dRing, func(i, j int) bool {
		return dRing[i].hash.Cmp(dRing[j].hash) == -1
	})

	if !testRingIntegrity(a) {
		t.Error("Failed ring integrity check after remove only node")
	}
	if a.Nodes.Length != config.NUM_VIRTUAL_NODES {
		t.Error("Deleting only node does not correctly reduce length")
	}
	if _, ok := a.NodeMap[names[0]]; ok {
		t.Error("Deleting node fails to clear info from NodeMap")
	}
	curr = a.Nodes.Head
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

	a.RemoveNode(names[1])
	if a.Nodes.Length != 0 {
		t.Error("Deleting remaining node does not correctly reduce length")
	}
	if a.Nodes.Head != nil {
		t.Error("Deleting remaining node does not correctly clear head of linked list")
	}
	if _, ok := a.NodeMap[names[1]]; ok {
		t.Error("Deleting node fails to clear info from NodeMap")
	}
}

func TestFindingKeys(t *testing.T) {
	if err := config.LoadEnvFile(); err != nil {
		t.Errorf("Error loading .env file: %v\n", err.Error())
	}
	keys := []string{"abc", "fdsafd", "hello", "1232187"}
	names := []string{"A", "B"}
	ip := []uint64{8000, 8001}
	a := NewRing()
	dRing := debugRing{}

	for i, name := range names {
		a.AddNode(name, ip[i])
		for j := 0; j < config.NUM_VIRTUAL_NODES; j++ {
			nodeName := fmt.Sprintf("%s_%d", name, j)
			dRing = append(dRing, debugNode{
				id:   nodeName,
				hash: Hash(nodeName),
			})
		}
	}
	if !testRingIntegrity(a) {
		t.Error("Failed ring integrity check")
	}

	sort.Slice(dRing, func(i, j int) bool {
		return dRing[i].hash.Cmp(dRing[j].hash) == -1
	})

	for _, key := range keys {
		hashedKey := Hash(key)
		j := sort.Search(len(names)*config.NUM_VIRTUAL_NODES, func(i int) bool {
			return dRing[i].hash.Cmp(hashedKey) != -1
		})
		if j == len(names)*config.NUM_VIRTUAL_NODES || j == 0 {
			j = len(names)*config.NUM_VIRTUAL_NODES - 1
		} else if j > 0 {
			j--
		}
		if dRing[j].id != a.SearchKey(key).VirtualName {
			t.Errorf("Failed key search for %s; expected %s but got %s", key, dRing[j].id, a.SearchKey(key).VirtualName)
		}
	}
}
