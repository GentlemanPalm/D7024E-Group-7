package d7024e

import (
	"fmt"
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

  //Replace with nodelookup
  kClosest := network.routingTable.FindClosestContacts(id, GetGlobals().K)
  rep := network.storeTable.GetNodesForRepublishing()
  fmt.Println("------Republishing------")
  fmt.Println(rep)
  fmt.Println("------------------------")
  for k,v := range rep { 
    for i := range kClosest { 
      network.SendStoreMessage(network.CreateStoreMessage(k, v, false), kClosest[i].Address)
    }  
  }
  
}