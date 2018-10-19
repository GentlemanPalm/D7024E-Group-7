package d7024e

import "testing"

func TestNodeLookupTable(t *testing.T) {
	table := NewNodeLookupTable()
	me := NewContact(NewRandomKademliaID(), "0.0.0.0")
	target := NewRandomKademliaID()
	shortlist := NewShortlist(&me, target, nil, 0)
	contact1 := NewContact(NewRandomKademliaID(), "0.0.0.1")
	contact2 := NewContact(NewRandomKademliaID(), "0.0.0.2")
	contacts := []Contact{contact1, contact2}
	shortlist.AddContacts(contacts)
	rID := NewRandomKademliaID()
	table.Put(rID, shortlist)
	if table.Pop(rID) == nil {
		t.Error("Should have item in NodeLookupTable")
	}
	if table.Pop(NewRandomKademliaID()) != nil {
		t.Error("Should not have some random item in table")
	}

}
