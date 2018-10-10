package d7024e

import (
	"fmt"
	"time"
)

func (network *Network) Republish(id *KademliaID) {
	//20 sekunder för demo.
	for {
		time.Sleep(12 * time.Second)
		go network.republish(id)
	}
}

func (network *Network) republish(id *KademliaID) {

	//Replace with nodelookup
	rep := network.storeTable.GetNodesForRepublishing()
	fmt.Println("------Republishing------")
	fmt.Println(rep)
	fmt.Println("------------------------")
	for k, v := range rep {
		network.SendStoreIterative(k, v)
	}

}
