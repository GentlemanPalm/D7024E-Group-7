package d7024e

import (
	"NetworkMessage"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

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

func TestSendPing(t *testing.T) {
	network := mockNetwork()
	mdw := &MockDataWriter{}
	network.dw = mdw

	contact := NewContact(NewRandomKademliaID(), "0.0.0.0")

	network.SendPingMessage(&contact)

	if mdw.nrofTimesCalled != 1 {
		t.Error("Receiving a ping did NOT result in a pong message being sent")
	} else {
		fmt.Println("Ping messages results in pongs")
	}

	randomId := NewRandomKademliaID()
	network.pingTable.Push(randomId, network.Me().ID, nil, nil)
	network.HandlePingTimeout(randomId, network.Me(), network.Me())
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
	network.pingTable.Push(id, kid, nil, nil)
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

func TestSendPongMessage(t *testing.T) {
	network := mockNetwork()
	mdw := &MockDataWriter{}
	sentPacket := &PacketContainer{}
	mdw.callback = sentPacket.findNodeCallback
	network.dw = mdw

	key := NewRandomKademliaID()

	pong := &NetworkMessage.Pong{
		RandomId:   key.String(),
		KademliaId: network.routingTable.Me().ID.String(),
		Address:    network.routingTable.Me().Address,
	}

	network.SendPongMessage(pong, "127.0.0.1")

	if mdw.nrofTimesCalled != 1 {
		t.Error("Expected to send Pong message after call to SendPongMessage!")
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

func TestFindValue(t *testing.T) {
	network := mockNetwork()
	mdw := &MockDataWriter{}
	sentPacket := &PacketContainer{}
	mdw.callback = sentPacket.findNodeCallback
	network.dw = mdw

	key := NewRandomKademliaID()
	recipient := NewContact(NewRandomKademliaID(), "127.0.0.1")

	network.SendFindDataMessage(key.String(), &recipient, nil, nil, nil)

	if mdw.nrofTimesCalled != 1 {
		t.Error("(FindValue) Expected packet to be sent")
	}
}

var findValueContent = []byte{1, 3, 3, 7}

type mockFileHandler struct {
	kvt map[string][]byte
}

func (mfh *mockFileHandler) ReadFile(hash string) []byte {
	return mfh.kvt[hash]
}

func (mfh *mockFileHandler) WriteFile(hash string, content []byte) bool {
	mfh.kvt[hash] = content
	return true
}

func TestHandleFindValue(t *testing.T) {
	network := mockNetwork()
	mfh := &mockFileHandler{}
	mfh.kvt = make(map[string][]byte)
	network.storeTable.fh = mfh

	mdw := &MockDataWriter{}
	sentPacket := &PacketContainer{}
	mdw.callback = sentPacket.findNodeCallback

	hash := NewRandomKademliaID()

	network.storeTable.Push(findValueContent, hash.String(), false, true)

	network.dw = mdw

	for i := 0; i < 25; i++ {
		contact := NewContact(NewRandomKademliaID(), "127.0.0."+strconv.Itoa(i))
		network.routingTable.AddContact(contact, network)
	}

	rid := NewRandomKademliaID()

	find := &NetworkMessage.Find{}
	find.RandomId = rid.String()
	find.Hash = hash.String()

	network.HandleFindValueMessage(find, "127.0.0.255")

	if mdw.nrofTimesCalled != 1 {
		t.Error("(HandleFindValue) Expected to have written a response to a findvalue message.")
	}

	if sentPacket.packet.Value == nil {
		t.Error("(HandleFindValue) No nodes were provided in response to findvalue.")
	} else {
		switch response := sentPacket.packet.Value.Response.(type) {
		case *NetworkMessage.ValueResponse_Nodes:
			t.Error("(HandleFindValue) Provided key known to be in store table. Exptected data response.")
		case *NetworkMessage.ValueResponse_Content:
			if response.Content == nil {
				t.Error("(HandleFindValue) Content was not the byte array as exptected")
			} else {
				fmt.Println("[OK] HandleFindValue handles requests known to contain data correctly.")
			}
		}
	}

	find = &NetworkMessage.Find{}
	rid = NewRandomKademliaID()
	hash = NewRandomKademliaID()

	find.RandomId = rid.String()
	find.Hash = hash.String()

	network.HandleFindValueMessage(find, "127.0.0.255")

	if sentPacket.packet.Value == nil {
		t.Error("(HandleFindValue) No nodes were provided in response to findvalue.")
	} else {
		switch response := sentPacket.packet.Value.Response.(type) {
		case *NetworkMessage.ValueResponse_Nodes:
			if len(response.Nodes.Nodes) != 20 {
				t.Errorf("(HandleFindValue) Response contains %d nodes, expected 20", len(response.Nodes.Nodes))
			} else {
				fmt.Println("HandleFindValue seems to work properly")
			}
		case *NetworkMessage.ValueResponse_Content:
			t.Error("There should not be a value returned for a random hash. There is an infinitesimal chance this is due to a collision.")
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
		network.routingTable.AddContact(contact, network)
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
		network.routingTable.AddContact(contact, network)
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

func TestValueLookup(t *testing.T) {
	network := mockNetwork()

	mfh := &mockFileHandler{}
	mfh.kvt = make(map[string][]byte)
	network.storeTable.fh = mfh

	mdw := &MockDataWriter{}
	sentPacket := &PacketContainer{}
	mdw.callback = sentPacket.findNodeCallback
	network.dw = mdw

	for i := 0; i < nrofContacts; i++ {
		contact := NewContact(NewRandomKademliaID(), "127.0.0."+strconv.Itoa(i))
		network.routingTable.AddContact(contact, network)
	}

	target := NewRandomKademliaID()

	network.storeTable.Push(findValueContent, target.String(), false, true)

	network.ValueLookup(target, onValueLookupFinish)

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
		responseID := sentPacket.packet.FindValue.RandomId
		nodes := &NetworkMessage.ValueResponse{
			RandomId: responseID,
			Response: &NetworkMessage.ValueResponse_Content{findValueContent},
			//&NetworkMessage.ValueResponse_Nodes{createNodeResponse(responseID, contacts)},
		}
		network.findTable.ProcessResult(nodes)
		//network.handleFindContactResponse(recipient.ID, nodes)
	}
	time.Sleep(time.Duration(7) * time.Second)

	network.ValueLookup(NewRandomKademliaID(), onValueLookupNotFound)

	for i := 0; i < nrofContacts; i++ {
		//time.Sleep(time.Duration(1) * time.Second / 2)
		responseID := sentPacket.packet.FindValue.RandomId
		nodes := &NetworkMessage.ValueResponse{
			RandomId: responseID,
			Response: &NetworkMessage.ValueResponse_Nodes{createNodeResponse(responseID, contacts)},
		}
		network.findTable.ProcessResult(nodes)
		//network.handleFindContactResponse(recipient.ID, nodes)
	}
	time.Sleep(time.Duration(15) * time.Second)
}

func onValueLookupFinish(contacts []Contact, data *[]byte) {
	if data == nil {
		fmt.Println("[ERR] (ValueLookup) Error: Data response expected for ValueLookup")
	} else {
		fmt.Println("[OK] Otherwise, ValueLookup seems OK.")
	}
}

func onValueLookupNotFound(contacts []Contact, data *[]byte) {
	missing := 0
	for i := range contacts {
		if contacts[i].ID == nil {
			missing++
		}
	}
	if data == nil {
		fmt.Println("[OK] (ValueLookup) node response is fine for value not found. " + strconv.Itoa(missing))
	} else {
		fmt.Println("[ERR] (ValueLookup) Should not receive data for random lookup")
	}
}

type PacketContainer struct {
	packet *NetworkMessage.Packet
}

func (pc *PacketContainer) findNodeCallback(packet *NetworkMessage.Packet) {
	pc.packet = packet
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
	store := network.CreateStoreMessage(hash, content, true)

	fmt.Println("File contents: %s", content)

	store.RandomId = NewRandomKademliaID().String()
	store.KademliaId = NewRandomKademliaID().String()
	store.Address = "127.0.0.1"
	network.SendStoreMessage(store, store.Address)
	network.HandleStoreMessage(store)
	fmt.Println(mdw.nrofTimesCalled)
	if mdw.nrofTimesCalled != 2 {
		t.Error("Receiving a store did NOT result in a response message being sent")
	} else {
		fmt.Println("Stor messages results in storeResponse")
	}
	k := NewRandomKademliaID()
	network.rpcTable.Push(k)
	response := network.CreateStoreResponseMessage(k)
	response.KademliaId = k.String()
	response.Address = "127.0.0.1"
	network.SendStoreResponseMessage(response, response.Address)
	network.HandleStoreResponseMessage(response)
}

func TestStoreIterativePalm(t *testing.T) {
	network := mockNetwork()

	mfh := &mockFileHandler{}
	mfh.kvt = make(map[string][]byte)
	network.storeTable.fh = mfh

	mdw := &MockDataWriter{}
	sentPacket := &PacketContainer{}
	mdw.callback = sentPacket.findNodeCallback
	network.dw = mdw

	randomId := NewRandomKademliaID()
	network.SendStoreIterative(randomId.String(), []byte("Hello, World"))

	hvc := &hashValueContainer{}
	hvc.network = network
	hvc.hash = randomId.String()
	hvc.content = []byte("Hello, World")

	contacts := make([]Contact, 20)
	for i := 0; i < len(contacts); i++ {
		contacts[i] = NewContact(NewRandomKademliaID(), "127.0.0."+strconv.Itoa(i))
	}

	data := []byte("Herp derp")
	hvc.storeOnReceivedContacts(contacts, &data)
}

func TestHandleReceive(t *testing.T) {
	network := mockNetwork()

	mfh := &mockFileHandler{}
	mfh.kvt = make(map[string][]byte)
	network.storeTable.fh = mfh

	mdw := &MockDataWriter{}
	sentPacket := &PacketContainer{}
	mdw.callback = sentPacket.findNodeCallback
	network.dw = mdw

	packet := network.createPacket()

	randomId := NewRandomKademliaID()
	randomID := randomId
	key := NewRandomKademliaID()

	packet.Ping = network.createPingPacket(randomId)
	packet.Pong = network.CreatePongMessage(packet.Ping)

	packet.FindNode = &NetworkMessage.Find{
		RandomId: randomID.String(),
		Hash:     key.String(),
	}

	contacts := make([]Contact, 20)

	for i := 0; i < 20; i++ {
		contacts[i] = NewContact(NewRandomKademliaID(), "0.0.0."+strconv.Itoa(i))
	}

	packet.Nodes = &NetworkMessage.ValueResponse{
		RandomId: randomID.String(),
		Response: &NetworkMessage.ValueResponse_Nodes{createNodeResponse(randomId.String(), contacts)},
	}

	packet.Value = &NetworkMessage.ValueResponse{
		RandomId: randomID.String(),
		Response: &NetworkMessage.ValueResponse_Nodes{createNodeResponse(randomId.String(), contacts)},
	}

	packet.Store = &NetworkMessage.Store{
		RandomId:   randomID.String(),
		KademliaId: network.routingTable.Me().ID.String(),
		Address:    network.routingTable.Me().Address,
		Hash:       key.String(),
		Content:    []byte("Derp"),
		Pin:        false,
	}

	/*packet.StoreResponse = &NetworkMessage.StoreResponse{
		RandomId:   randomID.String(),
		KademliaId: network.routingTable.Me().ID.String(),
		Address:    network.routingTable.Me().Address,
	}*/

	packet.FindValue = &NetworkMessage.Find{
		RandomId: randomID.String(),
		Hash:     key.String(),
	}

	null := network.createPacket()

	network.processPacket(packet)
	network.processPacket(null)
	iaddr := getIaddr()
	fmt.Println(iaddr)
}

func TestNetworkListen(t *testing.T) {
	me := NewContact(NewRandomKademliaID(), "127.0.0.1")
	routingTable := NewRoutingTable(me)
	network := NewNetwork(routingTable)
	//go network.Listen(42042)
	time.Sleep(time.Duration(3) * time.Second)

	network.SendPingMessage(&me)

	invalid := NewContact(NewRandomKademliaID(), "127.0.0.0")
	network.SendPingMessage(&invalid)

	time.Sleep(time.Duration(3) * time.Second)
}

func TestHandleFindDataResponse(t *testing.T) {
	randomId := NewRandomKademliaID()
	contacts := make([]Contact, 20)

	for i := 0; i < 20; i++ {
		contacts[i] = NewContact(NewRandomKademliaID(), "0.0.0."+strconv.Itoa(i))
	}
	nodes := &NetworkMessage.ValueResponse{
		RandomId: randomId.String(),
		Response: &NetworkMessage.ValueResponse_Nodes{createNodeResponse(randomId.String(), contacts)},
	}

	value := &NetworkMessage.ValueResponse{
		RandomId: randomId.String(),
		Response: &NetworkMessage.ValueResponse_Nodes{createNodeResponse(randomId.String(), contacts)},
	}
	cbc := &callbackContainer{}
	me := NewContact(NewRandomKademliaID(), "127.0.0.1")
	cbc.me = &me
	cbc.onContacts = cbcTestContacts
	cbc.onData = cbcTestData
	cbc.handleFindDataResponse(randomId, nodes)
	cbc.handleFindDataResponse(randomId, value)
}

func cbcTestContacts(id *KademliaID, contacts []Contact) {

}
func cbcTestData(id *KademliaID, data []byte) {

}
