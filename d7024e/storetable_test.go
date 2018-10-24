package d7024e

import (
	"testing"

	"math/rand"
	"time"
	"log"
	"io/ioutil"

)

func mockNet() *Network {
	rand.Seed(int64(time.Now().Nanosecond()))
	me := NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000")
	routingTable := NewRoutingTable(me)
	network := NewNetwork(routingTable)
	return network
}

func TestStoreTable(t *testing.T) {
	network := mockNet()

	fileName := Hash("text.txt")
	fileName2 := Hash("text2.txt")

	content, err1 := ioutil.ReadFile("text.txt")
	if err1 != nil {
		log.Fatal(err1)
	}

	PushRes := network.storeTable.Push(content, fileName, true, true)
	if !PushRes{
		t.Error("Fisrt file did not push to storeTable")
	}

	PushRes2 := network.storeTable.Push(content, fileName, true, true)
	if !PushRes2{
		t.Error("Second file did not push to storeTable")
	}

	Getres := network.storeTable.Get(fileName)
	if Getres == nil {
		t.Error("No returned file")
	}
	Getres2 := network.storeTable.Get(fileName2)
	if Getres2 != nil {
		t.Error("Failed get nil 2")
	}

	Pinres := network.storeTable.Pin(fileName)
	if Pinres == false {
		t.Error("NO PINNED FILE")
	}
	Pinres2 := network.storeTable.Pin(fileName2)
	if Pinres2 != false {
		t.Error("Failed Pin 2")
	}

	nodes := network.storeTable.GetNodesForRepublishing()
	if nodes == nil {
		t.Error("No nodes for republish")
	}

	Unpinres := network.storeTable.Unpin(fileName)
	if Unpinres == false {
		t.Error("NODE NOT UNPINNED")
	}
	Unpinres2 := network.storeTable.Unpin(fileName2)
	if Unpinres2 != false {
		t.Error("Failed unPin 2")
	}

	DeleteRes := network.storeTable.Delete(fileName)
	if DeleteRes == nil {
		t.Error("No Success delete")
	}

}

