package d7024e

import (
	"testing"

)

func TestPingTable(t *testing.T) {
	table := NewPingTable()
	rID := NewRandomKademliaID()
	kID := NewRandomKademliaID()
	notFound := NewRandomKademliaID()
	table.Push(rID, kID,nil,nil)
	elem := table.Pop(rID)
	if elem == nil {
		t.Error("Expected to find element added to table")
	}
	elem = table.Pop(notFound)
	if elem != nil {
		t.Error("Did not expect to find 'notFound' in table")
	}
}
