package d7024e

import (
	//"fmt"
	"time"
)

func (network *Network) Republish(id *KademliaID) {
	//20 sekunder för demo.
	for {
    	time.Sleep(30 * time.Second)
    	go network.republish(id)
  }
}

func (network *Network) republish(id *KademliaID) {

  kClosest := network.routingTable.FindClosestContacts(id, GetGlobals().K)
  for i := range kClosest { 
  	//HArd coded
  	network.SendStoreMessage(network.CreateStoreMessage("d7024e/text.txt"), kClosest[i].Address)
  }
}