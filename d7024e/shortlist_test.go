package d7024e

import (
	"NetworkMessage"
	"strconv"
	"testing"
)

func TestPrune(t *testing.T) {
	me := NewContact(NewRandomKademliaID(), "0.0.0.0")
	target := NewRandomKademliaID()
	shortlist1 := NewShortlist(&me, target, nil, 0)
	contacts := make([]Contact, 20)
	for i := 0; i < 20; i++ {
		contacts[i] = NewContact(NewRandomKademliaID(), "0.0.0."+strconv.Itoa(i))
	}
	contact1 := NewContact(NewRandomKademliaID(), "0.0.0.254")
	shortlist1.AddContacts(contacts)
	shortlist1.AddContacts([]Contact{contact1})
	if len(shortlist1.items) > 20 {
		t.Error("Should only have k elements in contact list")
	}
}

func TestGetClosestUnvisited(t *testing.T) {
	me := NewContact(NewRandomKademliaID(), "0.0.0.0")
	target := NewRandomKademliaID()
	shortlist1 := NewShortlist(&me, target, nil, 0)
	contact1 := NewContact(NewRandomKademliaID(), "0.0.0.1")
	contact2 := NewContact(NewRandomKademliaID(), "0.0.0.2")
	contacts := []Contact{contact1, contact2}
	shortlist1.AddContacts(contacts)

	closest1 := shortlist1.GetClosestUnvisited()
	if !(closest1.ID.Equals(contact1.ID) || closest1.ID.Equals(contact2.ID)) {
		t.Error("Expected to find contact among closest unvisited")
	}

	closest2 := shortlist1.GetClosestUnvisited()
	if closest2.ID.Equals(closest1.ID) {
		t.Error("Expected closest2 to not be same as closest1, as closest 1 should be marked as visited by now.")
	}

	if !(closest2.ID.Equals(contact1.ID) || closest2.ID.Equals(contact2.ID)) {
		t.Error("Expected to find contact among closest unvisited")
	}

	closest3 := shortlist1.GetClosestUnvisited()
	if closest3 != nil {
		t.Error("There should be nothing else in the shortlist")
	}
}

func TestMarkAsDead(t *testing.T) {
	me := NewContact(NewRandomKademliaID(), "0.0.0.0")
	target := NewRandomKademliaID()
	shortlist1 := NewShortlist(&me, target, nil, 0)
	contact1 := NewContact(NewRandomKademliaID(), "0.0.0.1")
	contact2 := NewContact(NewRandomKademliaID(), "0.0.0.2")
	contacts := []Contact{contact1, contact2}
	shortlist1.AddContacts(contacts)

	closest1 := shortlist1.GetClosestUnvisited()
	if !(closest1.ID.Equals(contact1.ID) || closest1.ID.Equals(contact2.ID)) {
		t.Error("Expected to find contact among closest unvisited")
	}

	shortlist1.MarkAsDead(&contact1)
	shortlist1.MarkAsDead(&contact2)

	closest2 := shortlist1.GetClosestUnvisited()
	if closest2 != nil {
		t.Error("Expected closest2 to be dead by now.")
	}
}

func TestEmptyLaunchRequests(t *testing.T) {
	me := NewContact(NewRandomKademliaID(), "0.0.0.0")
	target := NewRandomKademliaID()
	shortlist1 := NewShortlist(&me, target, nil, 0)

	network := mockNetwork()
	mdw := &MockDataWriter{}
	sentPacket := &PacketContainer{}
	mdw.callback = sentPacket.findNodeCallback
	network.dw = mdw

	if shortlist1.LaunchRequests(network) {
		t.Error("Empty contact list should result in failed LaunchRequests")
	}
}

func TestHandleShortlistResponse(t *testing.T) {
	me := NewContact(NewRandomKademliaID(), "0.0.0.0")
	target := NewRandomKademliaID()
	shortlist1 := NewShortlist(&me, target, nil, 0)
	contact1 := NewContact(NewRandomKademliaID(), "0.0.0.1")
	contact2 := NewContact(NewRandomKademliaID(), "0.0.0.2")
	//contact3 := NewContact(NewRandomKademliaID(), "0.0.0.3")
	contacts := []Contact{contact2}
	shortlist1.AddContacts(contacts)

	network := mockNetwork()
	mdw := &MockDataWriter{}
	sentPacket := &PacketContainer{}
	mdw.callback = sentPacket.findNodeCallback
	network.dw = mdw

	vRes := &NetworkMessage.ValueResponse{
		RandomId: contact2.ID.String(),
		Response: &NetworkMessage.ValueResponse_Nodes{createNodeResponse(contact2.ID.String(), []Contact{contact1})},
	}

	shortlist1.HandleResponse(network, contact2.ID, vRes)
	shortlist1.MarkAsDead(&contact2)

	close1 := shortlist1.GetClosestUnvisited()
	if close1 != nil {
		t.Error("Should not have a contact after receiving response")
	}
}

func TestHandleShortlistResponseValue(t *testing.T) {
	me := NewContact(NewRandomKademliaID(), "0.0.0.0")
	target := NewRandomKademliaID()
	shortlist1 := NewShortlist(&me, target, nil, 0)
	//contact1 := NewContact(NewRandomKademliaID(), "0.0.0.1")
	contact2 := NewContact(NewRandomKademliaID(), "0.0.0.2")
	//contact3 := NewContact(NewRandomKademliaID(), "0.0.0.3")
	contacts := []Contact{contact2}
	shortlist1.AddContacts(contacts)
	shortlist1.lookupValue = 1

	network := mockNetwork()
	mdw := &MockDataWriter{}
	sentPacket := &PacketContainer{}
	mdw.callback = sentPacket.findNodeCallback
	network.dw = mdw

	value := []byte("Derp")

	vRes := &NetworkMessage.ValueResponse{
		RandomId: contact2.ID.String(),
		Response: &NetworkMessage.ValueResponse_Content{value},
	}

	shortlist1.HandleResponse(network, contact2.ID, vRes)
}

func TestHandleShortlistTimeout(t *testing.T) {
	me := NewContact(NewRandomKademliaID(), "0.0.0.0")
	target := NewRandomKademliaID()
	shortlist1 := NewShortlist(&me, target, nil, 0)
	//contact1 := NewContact(NewRandomKademliaID(), "0.0.0.1")
	contact2 := NewContact(NewRandomKademliaID(), "0.0.0.2")
	//contact3 := NewContact(NewRandomKademliaID(), "0.0.0.3")
	contacts := []Contact{contact2}
	shortlist1.AddContacts(contacts)
	shortlist1.lookupValue = 1

	network := mockNetwork()
	mdw := &MockDataWriter{}
	sentPacket := &PacketContainer{}
	mdw.callback = sentPacket.findNodeCallback
	network.dw = mdw

	shortlist1.HandleTimeout(network, contact2.ID)
	closest := shortlist1.GetClosestUnvisited()
	if closest != nil {
		t.Error("Timeout should have marked only unvisited contact as DEAD.")
	}

}
