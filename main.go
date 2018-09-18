package main

import (
	"d7024e"
	"fmt"
	"NetworkMessage"
)

func main() {
	var x *d7024e.KademliaID
	x = d7024e.NewRandomKademliaID()
	var y = NetworkMessage.SearchRequest{}
	fmt.Println("Hello, World! " + x.String() + y.String())
	for {
		//fmt.Println("hue")
	}
}
