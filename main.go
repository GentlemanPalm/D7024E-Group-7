package main

import (
	"fmt"
	"d7024e"
)

func main() {
	var x *d7024e.KademliaID
	x = d7024e.NewRandomKademliaID()
	fmt.Println("Hello, World! "+x.String());
}
