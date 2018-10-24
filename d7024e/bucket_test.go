package d7024e

import (
	"testing"

	"math/rand"
	"time"

)
func mockNets() *Network {
	rand.Seed(int64(time.Now().Nanosecond()))
	me := NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000")
	routingTable := NewRoutingTable(me)
	network := NewNetwork(routingTable)
	return network
}

func TestBucket(t *testing.T) {
	network := mockNets()  
	bucket := NewBucket()
	
	contact1 := NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8001")
	contact2 := NewContact(NewKademliaID("1111111100000000000000000000000000000000"), "localhost:8002")
	contact3 := NewContact(NewKademliaID("1111111200000000000000000000000000000000"), "localhost:8002")
	
	bucket.AddContact(contact1, network)
	bucket.AddContact(contact2, network)

	bucket.UpdateBucket(&contact2)
	bucket.ReplaceContact(contact1.ID,&contact3,network)

	randomID := NewRandomKademliaID()

	distance := bucket.GetContactAndCalcDistance(randomID)
	if distance == nil {
		t.Error("contacts within distance is nil")
	}

}
