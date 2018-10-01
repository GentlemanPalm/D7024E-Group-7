package d7024e

import (
	"fmt"
	"time"
)

func Republish(network *Network) {
	//30 sekunder för demo.
	for {
    	time.Sleep(10 * time.Second)
    	go republish(network)
  }
}

func republish(network *Network) {
  kClosest := network.routingTable.FindClosestContacts(network.routingTable.Me().ID, GetGlobals().K)
  fmt.Printf("\n" + "Tjoooohoooo!1")
  fmt.Printf("\n")
  fmt.Printf("%v", kClosest)
}