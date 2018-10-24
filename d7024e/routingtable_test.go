package d7024e

import (
	"fmt"
	"testing"

	"math/rand"
	"time"
)

func mockN() *Network {
	rand.Seed(int64(time.Now().Nanosecond()))
	me := NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000")
	routingTable := NewRoutingTable(me)
	network := NewNetwork(routingTable)
	return network
}

func TestRoutingTable(t *testing.T) {
	network := mockN()

	kademliaid:= NewKademliaID("1111111400000000000000000000000000000000")
	con := NewContact(kademliaid, "localhost:8002")
	con2 := NewContact(NewKademliaID("1111111500000000000000000000000000000000"), "localhost:8002")

	network.routingTable.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8001"),network)
	network.routingTable.AddContact(NewContact(NewKademliaID("1111111100000000000000000000000000000000"), "localhost:8002"),network)
	network.routingTable.AddContact(NewContact(NewKademliaID("1111111200000000000000000000000000000000"), "localhost:8002"),network)
	network.routingTable.AddContact(NewContact(NewKademliaID("1111111300000000000000000000000000000000"), "localhost:8002"),network)
	network.routingTable.AddContact(con,network)
	network.routingTable.AddContact(NewContact(NewKademliaID("2111111400000000000000000000000000000000"), "localhost:8002"),network)



	network.routingTable.ReplaceContact(kademliaid , &con2 , network)
	network.routingTable.UpdateBucket(&con)

	contacts := network.routingTable.FindClosestContacts(NewKademliaID("2111111400000000000000000000000000000000"), 20)
	for i := range contacts {
		fmt.Println(contacts[i].String())
	}
}
