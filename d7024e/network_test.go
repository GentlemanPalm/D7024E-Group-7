package d7024e

import (
	"NetworkMessage"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"
	"io/ioutil"
	"log"
	"os"
	"github.com/golang/protobuf/proto"
)

func mockNetwork() *Network {
	rand.Seed(int64(time.Now().Nanosecond()))
	me := NewContact(NewRandomKademliaID(), getIaddr())
	routingTable := NewRoutingTable(me)
	network := NewNetwork(routingTable)
	return network
}

type PacketCallback func(*NetworkMessage.Packet)

type MockDataWriter struct {
	nrofTimesCalled int
	callback        PacketCallback
}

func (mdw *MockDataWriter) sendDataToAddress(address string, data []byte) {
	mdw.nrofTimesCalled += 1
	if mdw.callback != nil {
		packet := &NetworkMessage.Packet{}
		err := proto.Unmarshal(data, packet)
		if err != nil {
			fmt.Println("UNMARSHAL FAIL")
			fmt.Println(err)
		}
		mdw.callback(packet)
	}

}

func TestHandleOrigin(t *testing.T) {
	network := mockNetwork()
	id := NewRandomKademliaID()
	closest := network.routingTable.FindClosestContacts(id, 20)
	if len(closest) != 0 {
		t.Error("Somehow the routing table is already populated.")
	}
	pair := &NetworkMessage.KademliaPair{}
	pair.KademliaId = id.String()
	pair.Address = "123.4.5.6"
	network.HandleOriginMessage(pair)
	closest = network.routingTable.FindClosestContacts(id, 20)
	if len(closest) != 1 {
		t.Error("Does not add item to routing table when Origin is received")
	} else {
		fmt.Println("HandleOrigin works as intended")
	}
}

func TestHandlePing(t *testing.T) {
	network := mockNetwork()
	mdw := &MockDataWriter{}
	network.dw = mdw
	ping := &NetworkMessage.Ping{}
	ping.RandomId = NewRandomKademliaID().String()
	ping.KademliaId = NewRandomKademliaID().String()
	ping.Address = "127.0.0.1"
	network.HandlePingMessage(ping)
	if mdw.nrofTimesCalled != 1 {
		t.Error("Receiving a ping did NOT result in a pong message being sent")
	} else {
		fmt.Println("Ping messages results in pongs")
	}
}

func TestHandlePong(t *testing.T) {
	network := mockNetwork()
	id := NewRandomKademliaID()
	kid := NewRandomKademliaID()
	closest := network.routingTable.FindClosestContacts(id, 20)
	if len(closest) != 0 {
		t.Error("(TestPong) Somehow the routing table is already populated.")
	}
	network.pingTable.Push(id, kid)
	pong := &NetworkMessage.Pong{}
	pong.RandomId = id.String()
	pong.KademliaId = kid.String()
	pong.Address = "127.0.0.1"
	network.HandlePongMessage(pong)
	closest = network.routingTable.FindClosestContacts(id, 20)
	if len(closest) != 1 {
		t.Error("(TestPong) Does not update routingTable upon successful pong")
	} else {
		fmt.Println("Pong is handled correctly")
	}

}

func TestFindNode(t *testing.T) {
	network := mockNetwork()
	mdw := &MockDataWriter{}
	sentPacket := &PacketContainer{}
	mdw.callback = sentPacket.findNodeCallback
	network.dw = mdw

	key := NewRandomKademliaID()
	recipient := NewContact(NewRandomKademliaID(), "127.0.0.1")

	network.SendFindContactMessage(key, &recipient)

	if mdw.nrofTimesCalled != 1 {
		t.Error("(FindNode) Expected packet to be sent")
	}

	if sentPacket.packet.FindNode == nil {
		t.Error("(FindNode) Exptected findNode field in sent packet. None was found")
	} else {
		fmt.Println("FindNode at least sends the packet!")
		contacts := make([]Contact, 25)
		for i := 0; i < len(contacts); i++ {
			contacts[i] = NewContact(NewRandomKademliaID(), "127.0.0."+strconv.Itoa(i))
		}
		nodes := &NetworkMessage.ValueResponse{
			RandomId: sentPacket.packet.FindNode.RandomId,
			Response: &NetworkMessage.ValueResponse_Nodes{createNodeResponse(sentPacket.packet.FindNode.RandomId, contacts)},
		}
		network.handleFindContactResponse(recipient.ID, nodes)
		closest := network.routingTable.FindClosestContacts(key, 20)
		if len(closest) < 20 {
			t.Errorf("(FindNode) Did not have 20 contats after response")
		} else {
			fmt.Println("FindNode seems to be working properly")
		}
	}

}

func TestHandleFindNode(t *testing.T) {
	network := mockNetwork()
	mdw := &MockDataWriter{}
	sentPacket := &PacketContainer{}
	mdw.callback = sentPacket.findNodeCallback
	network.dw = mdw

	for i := 0; i < 25; i++ {
		contact := NewContact(NewRandomKademliaID(), "127.0.0."+strconv.Itoa(i))
		network.routingTable.AddContact(contact)
	}

	rid := NewRandomKademliaID()
	hash := NewRandomKademliaID()

	find := &NetworkMessage.Find{}
	find.RandomId = rid.String()
	find.Hash = hash.String()

	network.HandleFindContactMessage(find, "127.0.0.255")

	if mdw.nrofTimesCalled != 1 {
		t.Error("(HandleFindNode) Expected to have written a response to a findcontact message.")
	}

	if sentPacket.packet.Nodes == nil {
		t.Error("(HandleFindNode) No nodes were provided in response to findnode.")
	} else {
		switch response := sentPacket.packet.Nodes.Response.(type) {
		case *NetworkMessage.ValueResponse_Nodes:
			if len(response.Nodes.Nodes) != 20 {
				t.Errorf("(HandleFindNode) Response contains %d nodes, expected 20", len(response.Nodes.Nodes))
			} else {
				fmt.Println("HandleFindNode seems to work properly")
			}
		case *NetworkMessage.ValueResponse_Content:
			t.Error("(HandleFindNode) Content response was detected for FindNode, which is undesirable")
		}
	}
}

const nrofContacts = 10

func TestNodeLookup(t *testing.T) {
	network := mockNetwork()
	mdw := &MockDataWriter{}
	sentPacket := &PacketContainer{}
	mdw.callback = sentPacket.findNodeCallback
	network.dw = mdw

	for i := 0; i < nrofContacts; i++ {
		contact := NewContact(NewRandomKademliaID(), "127.0.0."+strconv.Itoa(i))
		network.routingTable.AddContact(contact)
	}

	target := NewRandomKademliaID()

	network.NodeLookup(target, onNodeLookupFinish)

	//time.Sleep(time.Duration(1) * time.Second)

	//key := sentPacket.packet.FindNode.Hash

	contacts := make([]Contact, 20)
	for i := 0; i < len(contacts); i++ {
		contacts[i] = NewContact(NewRandomKademliaID(), "127.0.0."+strconv.Itoa(i))
	}

	if mdw.nrofTimesCalled != 3 { // TODO: Use alpha
		t.Errorf("(NodeLookup) Error: Expected 3 packets sent from NodeLookup, received %d", mdw.nrofTimesCalled)
	}

	for i := 0; i < nrofContacts; i++ {
		//time.Sleep(time.Duration(1) * time.Second / 2)
		responseID := sentPacket.packet.FindNode.RandomId
		nodes := &NetworkMessage.ValueResponse{
			RandomId: responseID,
			Response: &NetworkMessage.ValueResponse_Nodes{createNodeResponse(responseID, contacts)},
		}
		network.findTable.ProcessResult(nodes)
		//network.handleFindContactResponse(recipient.ID, nodes)
	}
	time.Sleep(time.Duration(7) * time.Second)
}

func onNodeLookupFinish(contacts []Contact, data *[]byte) {
	missing := 0
	for i := range contacts {
		if contacts[i].ID == nil {
			missing++
		}
	}
	if data != nil {
		fmt.Println("[ERR] (NodeLookup) Error: Data response received for NodeLookup")
	} else {
		fmt.Println("[OK] Otherwise, NodeLookup seems OK.")
	}
}

type PacketContainer struct {
	packet *NetworkMessage.Packet
}

func (pc *PacketContainer) findNodeCallback(packet *NetworkMessage.Packet) {
	pc.packet = packet
}

func TestFindValue(t *testing.T) {
	// TODO: After FIND_VALUE is implemented and Test_FIND_NODE works
}
//Store RPC
func TestStore(t *testing.T) {
	network := mockNetwork()
	mdw := &MockDataWriter{}
	network.dw = mdw

	filename := "text.txt"
	hash := Hash(filename)

	os.Mkdir("Files", 0644)
	filePath := "Files/" + hash 
	content, err1 := ioutil.ReadFile(filename)
		if err1 != nil {
			log.Fatal(err1)
	}

	err2 := ioutil.WriteFile(filePath, content, 0644)
	if err2 != nil {
		log.Fatal(err2)
	}
	//Test
	store := &NetworkMessage.Store{}

	fmt.Println("File contents: %s", content)

	store.RandomId = NewRandomKademliaID().String()
	store.KademliaId = NewRandomKademliaID().String()
	store.Address = "127.0.0.1"
	store.Hash = hash
	store.Content = content
	store.Pin = true
	network.HandleStoreMessage(store)
	if mdw.nrofTimesCalled != 1 {
		t.Error("Receiving a store did NOT result in a pong message being sent")
	} else {
		fmt.Println("Stor messages results in storeResponse")
	}
}
/*func TestHandleStore(t *testing.T) {
	network := mockNetwork()
	mdw := &MockDataWriter{}
	sentPacket := &PacketContainer{}
	mdw.callback = sentPacket.findNodeCallback
	network.dw = mdw

	
}*/


