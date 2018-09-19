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
	
	//Test for hash
	var str = "d7024e/text.txt"
	fmt.Println("\n" + "Hello, I hashed this file: ");
	d7024e.Hash(str)

	for {
		//fmt.Println("hue")
	}
}
