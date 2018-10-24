package d7024e

import (
	"testing"
)

func TestRpcTable(t *testing.T) {
	table := NewRpcTable()
	rID := NewRandomKademliaID()
	table.Push(rID)
	elem := table.Pop(rID)
	if elem == nil {
		t.Error("Expected to find element added to table")
	}
}