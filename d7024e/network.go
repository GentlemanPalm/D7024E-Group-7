package d7024e

import (
	"NetworkMessage"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
)

type Network struct {
	routingTable    *RoutingTable
	pingTable       *PingTable
	findTable       *FindTable
	nodeLookupTable *NodeLookupTable
	dw              DataWriter
	rpcTable        *RpcTable
	storeTable      *StoreTable
}

func NewNetwork(routingTable *RoutingTable) *Network {
	nw := &Network{}
	nw.routingTable = routingTable
	nw.routingTable.Me().Address = getIaddr()
	nw.pingTable = NewPingTable() // TODO: Create dependency injection
	nw.findTable = NewFindTable()
	nw.nodeLookupTable = NewNodeLookupTable()
	nw.dw = &NetworkDataWriter{}
	nw.rpcTable = NewRpcTable()
	nw.storeTable = NewStoreTable()
	go nw.storeTable.Expire()
	return nw
}

func (network *Network) Me() *Contact {
	return network.routingTable.Me()
}

func (network *Network) GetStoreTable() *StoreTable {
	return network.storeTable
}

func (network *Network) Listen(port int) {
	// Taken almost directly from
	// http://www.minaandrawos.com/2016/05/14/udp-vs-tcp-in-golang/

	// listen to incoming udp packets
	//var nrofPacketsRcvd int = 0
	fmt.Println("Entered Network.Listen")

	//saddr, _ := net.ResolveUDPAddr("udp", ":"+defaultPort)

	pc, err := net.ListenPacket("udp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Fatal(err)
	}
	defer pc.Close()
	fmt.Println("Now listening for connections")

	for {
		//simple read

		fmt.Print("Reading from ListenPacket...")
		buffer := make([]byte, 8192)
		size, addr, err := pc.ReadFrom(buffer)
		go network.handleReceive(buffer, size, addr.String(), err)

		//simple write
		//pc.WriteTo([]byte("Hello from client"), net.ResolveUDPAddr("udp", ":2000"))
	}
}

func (network *Network) handleReceive(buffer []byte, size int, addr string, err error) {
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Received packet from " + addr)

	packet := &NetworkMessage.Packet{}
	marshalerr := proto.Unmarshal(buffer[:size], packet)
	if marshalerr != nil {
		fmt.Println("Received an error from the ping command")
		fmt.Println(marshalerr)
	}
	//	s := string(buffer[:14])
	if packet.Ping != nil {
		ping := packet.Ping
		fmt.Println("Received: " + ping.RandomId + " from " + addr)
	}

	network.processPacket(packet)
}

func (network *Network) processPacket(packet *NetworkMessage.Packet) {
	if packet.Origin != nil {
		network.HandleOriginMessage(packet.Origin) // Do synchronously to prevent race conditions
	} else {
		fmt.Println("Received packet, but ORIGIN left blank. Bug?")
	}

	if packet.Ping != nil {
		go network.HandlePingMessage(packet.Ping)
	} else {
		fmt.Println("Received packet, but PING left blank")
	}

	if packet.Pong != nil {
		go network.HandlePongMessage(packet.Pong)
	} else {
		fmt.Println("Received packet, but PONG left blank")
	}

	if packet.Store != nil {
		go network.HandleStoreMessage(packet.Store)
	} else {
		fmt.Println("Received packet, but Store left blank")
	}

	if packet.StoreResponse != nil {
		go network.HandleStoreResponseMessage(packet.StoreResponse)
	} else {
		fmt.Println("Received packet, but StoreResponse left blank")
	}

	if packet.FindNode != nil {
		if packet.Origin != nil {
			go network.HandleFindContactMessage(packet.FindNode, packet.Origin.Address)
		} else {
			fmt.Println("ERROR: Received FIND_NODE without ORIGIN. Does not know how to respond.")
		}
	} else {
		fmt.Println("Received packet, but FIND_NODE left blank")
	}

	if packet.Nodes != nil {
		go network.findTable.ProcessResult(packet.Nodes)
	} else {
		fmt.Println("Received packet, but FIND_NODE_RESPONSE left blank")
	}

	if packet.FindValue != nil {
		if packet.Origin != nil {
			go network.HandleFindValueMessage(packet.FindValue, packet.Origin.Address) // TODO: Implement
		} else {
			fmt.Println("Received packet, but FIND_VALUE was left blank")
		}
	}

	if packet.Value != nil {
		go network.findTable.ProcessResult(packet.Value)
	} else {
		fmt.Println("Received packet, but FIND_VALUE_RESPONSE was left blank")
	}

}

func (network *Network) HandleOriginMessage(origin *NetworkMessage.KademliaPair) {
	fmt.Println("Received an origin message")
	fmt.Println("id=" + origin.KademliaId + " addr=" + origin.Address)
	if !network.Me().ID.Equals(NewKademliaID(origin.KademliaId)) {
		fmt.Println("Added contact")
		network.routingTable.AddContact(NewContact(NewKademliaID(origin.KademliaId), origin.Address), network)
	} else {
		fmt.Println("Received origin message from self. Won't add.")
	}
	result := network.routingTable.FindClosestContacts(NewKademliaID(origin.KademliaId), 20)
	for i := range result {
		fmt.Println("Has close contact " + result[i].ID.String() + " at " + result[i].Address)
	}
}

// Yank function to determine IP on local network with docker.
// Might not work for outbound traffic

func getIaddr() string {
	ifaces, _ := net.Interfaces()
	// handle err
	iaddr := "127.0.0.1"
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
		// handle err
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip.String() != "127.0.0.1" {
				iaddr = ip.String()
			}
		}
	}
	return iaddr
}

func (network *Network) HandlePingMessage(pingMessage *NetworkMessage.Ping) {
	fmt.Println("Received Ping Message. I should update the buckets here at some point")
	contact := NewContact(NewKademliaID(pingMessage.KademliaId), pingMessage.Address)

	if !contact.ID.Equals(network.Me().ID) {
		network.routingTable.AddContact(contact,network)
		fmt.Println("Added " + pingMessage.KademliaId + " @ " + pingMessage.Address + " as a contact from ping")
	} else {
		fmt.Println("Received oneself as parameter to ping message. Decided against adding it to contact list.")
	}

	network.SendPongMessage(network.CreatePongMessage(pingMessage), pingMessage.Address)
}

func (network *Network) CreatePongMessage(pingMessage *NetworkMessage.Ping) *NetworkMessage.Pong {
	pong := &NetworkMessage.Pong{
		RandomId:   pingMessage.RandomId,
		KademliaId: network.routingTable.Me().ID.String(),
		Address:    network.routingTable.Me().Address,
	}
	return pong
}

func (network *Network) HandlePingTimeout(randomID *KademliaID,old *Contact, replacement *Contact) {
	time.Sleep(time.Duration(2) * time.Second)
	// Atomic operation, removes the item from the table and returns it
	row := network.pingTable.Pop(randomID)
	fmt.Println("-----POP FROM TABLE-------")
	fmt.Println(randomID)
	// Nil row implies a response was found in time
	if row != nil {
		fmt.Println("PingTable entry for " + randomID.String() + " remained after timeout.")
		if replacement == nil {
			fmt.Println("No replacement was found. Deleting.")
		}else {
			fmt.Println("The replacement has an ID of " + replacement.ID.String() + " TODO: IMPLEMENT")
			if row.onTimeout != nil{
				fmt.Println("-----------GOING FOR TIMEOUT----------")
				go row.onTimeout(old.ID, replacement, network)	
			}else{

			}
		}
	} else {
		fmt.Println("Looked to the table for " + randomID.String() + " but received a response in time")
	}
}

func (network *Network) HandlePongMessage(pongMessage *NetworkMessage.Pong) {
	// Atomically remove the item from the table and get the row
	fmt.Println("-----POP FROM TABLE-------")
	fmt.Println(NewKademliaID(pongMessage.RandomId))
	row := network.pingTable.Pop(NewKademliaID(pongMessage.RandomId))
	var contact Contact
	if row == nil {
		fmt.Println("Received pong with random id " + pongMessage.RandomId + " but nothing was found in the ping table")
		contact = NewContact(NewKademliaID(pongMessage.KademliaId), pongMessage.Address)
		if !network.Me().ID.Equals(contact.ID) {
			network.routingTable.AddContact(contact,network)
		} else {
			fmt.Println("Recevied pong from self. Not adding to contact list")
		}
	} else {
		contact = NewContact(row.kademliaID, pongMessage.Address)
		if row.onResponse != nil{
			fmt.Println("-----------GOING FOR RESPONSE----------")
			go row.onResponse(&contact)	
		}else{
			if !network.Me().ID.Equals(contact.ID) {
				fmt.Println("-----------GOING FOR ADD CONTACT----------")
				network.routingTable.AddContact(contact,network)
			} else {
				fmt.Println("Recevied pong from self. Not adding to contact list")
			}
		}
	}
	
	fmt.Println("Got the PONG message for " + pongMessage.KademliaId + " with random ID " + pongMessage.RandomId)
}

// Generic interface for writing data
type DataWriter interface {
	sendDataToAddress(string, []byte)
}

type NetworkDataWriter struct {
}

func (ndw *NetworkDataWriter) sendDataToAddress(address string, data []byte) {
	saddr, e0 := net.ResolveUDPAddr("udp", address)
	if e0 != nil {
		fmt.Println(e0)
	}
	conn, err := net.DialUDP("udp", nil, saddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	//rnum := rand.Int() % 100000
	//simple write
	fmt.Println("Trying to write a packet of data to " + address)

	_, werr := conn.Write(data)
	if werr != nil {
		fmt.Println("Something went wrong with sending data to " + address)
		fmt.Println(werr)
	}
	fmt.Println("Wrote a packet of data")
}

func (network *Network) createPacket() *NetworkMessage.Packet {
	packet := &NetworkMessage.Packet{}
	packet.Origin = &NetworkMessage.KademliaPair{network.Me().ID.String(), network.Me().Address}
	return packet
}

func ensurePort(address string, port string) string {
	adr := strings.Split(address, ":")
	return adr[0] + ":" + port
}

func (network *Network) SendPongMessage(pongMessage *NetworkMessage.Pong, address string) {
	packet := network.createPacket()
	packet.Pong = pongMessage
	out, merr := proto.Marshal(packet)
	if merr != nil {
		fmt.Println("Error marshaling ping packet")
	} else {
		//fmt.Println("Marshalled data is " + string(out[:]))
		fmt.Println("RandomId is " + pongMessage.RandomId)
		fmt.Println("Sending PONG to " + address)
		network.dw.sendDataToAddress(ensurePort(address, "42042"), out)
	}

}

// PING RPC
func (network *Network) SendPingMessage(contact *Contact) {
	network.SendPingMessageWithReplacement(contact, nil , nil , nil)
}

func (network *Network) SendPingMessageWithReplacement(contact *Contact, replacement *Contact, onTimeout func(*KademliaID, *Contact, *Network), onResponse func(*Contact)) {
	randomID := NewRandomKademliaID()
	fmt.Println("-----PUSH TO TABLE-------")
	fmt.Println(randomID)
	network.pingTable.Push(randomID, contact.ID,onTimeout,onResponse)
	go network.HandlePingTimeout(randomID,contact, replacement)
	network.sendPingPacket(randomID, contact)
}

func (network *Network) createPingPacket(randomID *KademliaID) *NetworkMessage.Ping {
	return &NetworkMessage.Ping{
		RandomId:   randomID.String(),
		KademliaId: network.routingTable.Me().ID.String(),
		Address:    network.routingTable.Me().Address,
	}
}

func (network *Network) sendPingPacket(randomID *KademliaID, contact *Contact) {
	packet := network.createPacket()
	packet.Ping = network.createPingPacket(randomID)

	out, merr := proto.Marshal(packet)
	if merr != nil {
		fmt.Println("Error marshalling ping packet")
	}

	network.dw.sendDataToAddress(ensurePort(contact.Address, "42042"), out)
}

// FIND_NODE RPC
func (network *Network) SendFindContactMessage(key *KademliaID, recipient *Contact) {
	network.sendFindContactMessage(key, recipient, nil, network.handleFindContactResponse)
}

func (network *Network) sendFindContactMessage(key *KademliaID, recipient *Contact, onTimeout func(*KademliaID, *KademliaID), onResponse func(*KademliaID, *NetworkMessage.ValueResponse)) {
	// TODO
	packet := network.createPacket()
	randomID := network.findTable.MakeRequest(recipient.ID, onTimeout, onResponse)

	packet.FindNode = &NetworkMessage.Find{
		RandomId: randomID.String(),
		Hash:     key.String(),
	}

	out, merr := proto.Marshal(packet)
	if merr != nil {
		fmt.Println("Error marshalling find_node packet")
	}

	network.dw.sendDataToAddress(ensurePort(recipient.Address, "42042"), out)
}

func (network *Network) handleFindContactResponse(recipient *KademliaID, message *NetworkMessage.ValueResponse) {
	fmt.Println("(hFCR) Received data from " + recipient.String() + " regarding " + message.RandomId)
	fmt.Println("Has been informed about the following nodes: ")
	switch response := message.Response.(type) {
	case *NetworkMessage.ValueResponse_Nodes:
		nodes := response.Nodes.Nodes
		for i := range nodes { // TODO: Make it work for FIND_VALUE
			fmt.Println(nodes[i].KademliaId + " @ " + nodes[i].Address)
			kID := NewKademliaID(nodes[i].KademliaId)
			if !network.Me().ID.Equals(kID) {
				network.routingTable.AddContact(NewContact(kID, nodes[i].Address), network)
			} else {
				fmt.Println("But that's me! I can't add myself now, can I?")
			}

		}
	case *NetworkMessage.ValueResponse_Content:
		fmt.Println("Cannot handle content values just yet")
	}

}

// This is for handling FindContactMessages sent from other machines
// This is NOT for handling the responses of requests sent by this node
func (network *Network) HandleFindContactMessage(findNode *NetworkMessage.Find, addr string) {
	//network.routingTable.AddContact(findNode.)
	fmt.Println("Received unsolicited find contact message. I should give a proper response instead of printing this message.")
	contacts := network.routingTable.FindClosestContacts(NewKademliaID(findNode.Hash), 20)
	packet := network.createPacket()
	packet.Nodes = &NetworkMessage.ValueResponse{
		RandomId: findNode.RandomId,
		Response: &NetworkMessage.ValueResponse_Nodes{createNodeResponse(findNode.RandomId, contacts)},
	}

	out, merr := proto.Marshal(packet)
	if merr != nil {
		fmt.Println("Error marshalling find_node RESPONSE packet")
	}

	network.dw.sendDataToAddress(ensurePort(addr, "42042"), out)
}

// This is for handling FindContactMessages sent from other machines
// This is NOT for handling the responses of requests sent by this node
func (network *Network) HandleFindValueMessage(findNode *NetworkMessage.Find, addr string) {
	//network.routingTable.AddContact(findNode.)
	fmt.Println("Received unsolicited find VALUE message. I should give a proper response instead of printing this message.")
	contacts := network.routingTable.FindClosestContacts(NewKademliaID(findNode.Hash), 20)
	packet := network.createPacket()

	value := network.storeTable.Get(findNode.Hash)
	if value == nil {
		packet.Value = &NetworkMessage.ValueResponse{
			RandomId: findNode.RandomId,
			Response: &NetworkMessage.ValueResponse_Nodes{createNodeResponse(findNode.RandomId, contacts)},
		}
	} else {
		fmt.Println("Actually found value with hash " + findNode.Hash + " in StoreTable.")
		packet.Value = &NetworkMessage.ValueResponse{
			RandomId: findNode.RandomId,
			Response: &NetworkMessage.ValueResponse_Content{value},
		}
	}

	out, merr := proto.Marshal(packet)
	if merr != nil {
		fmt.Println("Error marshalling find_value RESPONSE packet")
	}

	network.dw.sendDataToAddress(ensurePort(addr, "42042"), out)
}

func createNodeResponse(randomID string, contacts []Contact) *NetworkMessage.NodeResponse {
	response := &NetworkMessage.NodeResponse{}
	response.RandomId = randomID
	response.Nodes = make([]*NetworkMessage.KademliaPair, len(contacts))
	for i := range response.Nodes {
		response.Nodes[i] = createKademliaPair(&contacts[i])
	}
	return response
}

func createKademliaPair(contact *Contact) *NetworkMessage.KademliaPair {
	fmt.Println("Creating contact KademliaPair (" + contact.ID.String() + ", " + contact.Address + ")")
	return &NetworkMessage.KademliaPair{
		KademliaId: contact.ID.String(),
		Address:    contact.Address,
	}
}

type NodeLookupCallback func([]Contact, *[]byte)

func (network *Network) NodeLookup(id *KademliaID, onFinish NodeLookupCallback) {
	// TODO
	// 1. Create a shortlist with the three closest items of the id
	//     determine which element is the closest element
	fmt.Println("Entering nodelookup for " + id.String())
	closest := network.routingTable.FindClosestContacts(id, 3) // TODO: Use alpha
	for i := range closest {
		fmt.Println("Has contact: " + closest[i].ID.String())
	}
	shortlist := NewShortlist(network.Me(), id, onFinish, 0)
	shortlist.AddContacts(closest)

	// 2. Send FIND_NODE or FIND_VALUE requests to the three nodes
	//     mark the items as visited in the table.

	for i := 0; i < 3; i++ {
		contact := shortlist.GetClosestUnvisited()
		if contact != nil {
			fmt.Println("Sending initial message for node lookup of " + id.String() + " to " + contact.ID.String())
			network.sendFindNodeForNodeLookup(id, contact, shortlist, network.handleNodeLookupTimeout, network.handleNodeLookupResponse)
		} else {
			fmt.Println("One of the 'closest unvisited' was NIL")
		}
	}
}

func (network *Network) ValueLookup(id *KademliaID, onFinish NodeLookupCallback) {
	// TODO
	// 1. Create a shortlist with the three closest items of the id
	//     determine which element is the closest element
	fmt.Println("Entering value lookup for " + id.String())
	closest := network.routingTable.FindClosestContacts(id, 3) // TODO: Use alpha
	for i := range closest {
		fmt.Println("Has contact: " + closest[i].ID.String())
	}
	shortlist := NewShortlist(network.Me(), id, onFinish, 1)
	shortlist.AddContacts(closest)

	// 2. Send FIND_NODE or FIND_VALUE requests to the three nodes
	//     mark the items as visited in the table.

	for i := 0; i < 3; i++ {
		contact := shortlist.GetClosestUnvisited()
		if contact != nil {
			fmt.Println("Sending initial message for VALUE lookup of " + id.String() + " to " + contact.ID.String())
			network.SendFindValueForValueLookup(id, contact, shortlist)
		} else {
			fmt.Println("One of the 'closest unvisited' was NIL")
		}
	}
}

func (network *Network) SendFindNodeForNodeLookup(key *KademliaID, recipient *Contact, shortlist *Shortlist) {
	network.sendFindNodeForNodeLookup(key, recipient, shortlist, network.handleNodeLookupTimeout, network.handleNodeLookupResponse)
}

func (network *Network) sendFindNodeForNodeLookup(key *KademliaID, recipient *Contact, shortlist *Shortlist, onTimeout func(*KademliaID, *KademliaID), onResponse func(*KademliaID, *NetworkMessage.ValueResponse)) {
	packet := network.createPacket()
	randomID := network.findTable.MakeRequest(recipient.ID, onTimeout, onResponse)

	network.nodeLookupTable.Put(randomID, shortlist)

	packet.FindNode = &NetworkMessage.Find{
		RandomId: randomID.String(),
		Hash:     key.String(),
	}

	out, merr := proto.Marshal(packet)
	if merr != nil {
		fmt.Println("Error marshalling find_node packet")
	}

	network.dw.sendDataToAddress(ensurePort(recipient.Address, "42042"), out)
}

func (network *Network) SendFindValueForValueLookup(key *KademliaID, recipient *Contact, shortlist *Shortlist) {
	network.sendFindValueForValueLookup(key, recipient, shortlist)
}

// (hash string, recipient *Contact, onData OnDataCallback, onContacts OnContactsCallback, onTimeout onTimeout func(*KademliaID, *KademliaID)) {

func (network *Network) sendFindValueForValueLookup(key *KademliaID, recipient *Contact, shortlist *Shortlist) {
	packet := network.createPacket()
	randomID := network.findTable.MakeRequest(recipient.ID, network.handleNodeLookupTimeout, network.handleNodeLookupResponse)

	network.nodeLookupTable.Put(randomID, shortlist)

	packet.FindValue = &NetworkMessage.Find{
		RandomId: randomID.String(),
		Hash:     key.String(),
	}

	out, merr := proto.Marshal(packet)
	if merr != nil {
		fmt.Println("Error marshalling find_value packet for vlaue lookup")
	}

	network.dw.sendDataToAddress(ensurePort(recipient.Address, "42042"), out)
}

// onTimeout func(*KademliaID), onResponse func(*KademliaID, *NetworkMessage.ValueResponse))

/*
 * Handler for when the NodeLookup
 * */
func (network *Network) handleNodeLookupResponse(sender *KademliaID, message *NetworkMessage.ValueResponse) {
	shortlist := network.nodeLookupTable.Pop(NewKademliaID(message.RandomId))
	fmt.Println("Handline Node/Value lookup response from " + sender.String())
	if shortlist != nil {
		shortlist.HandleResponse(network, sender, message)
	} else {
		fmt.Println("Tried to handle the response from " + sender.String() + " but nodeLookupTable returned nil.")
	}

}

func (network *Network) handleNodeLookupTimeout(sender *KademliaID, randomID *KademliaID) {
	shortlist := network.nodeLookupTable.Pop(randomID)
	if shortlist != nil {
		shortlist.HandleTimeout(network, sender)
	} else {
		fmt.Println("Tried to handle timeout for shortlist, but nodeLookupTable returned nil!")
	}

}

type OnDataCallback func(*KademliaID, []byte)
type OnContactsCallback func(*KademliaID, []Contact)

type callbackContainer struct {
	onData     OnDataCallback
	onContacts OnContactsCallback
	me         *Contact
}

func (cbc *callbackContainer) handleFindDataResponse(sender *KademliaID, message *NetworkMessage.ValueResponse) {
	fmt.Println("(handleFindDataResponse) Received data from " + sender.String() + " regarding " + message.RandomId)
	switch response := message.Response.(type) {
	case *NetworkMessage.ValueResponse_Nodes:
		fmt.Println("Has been informed about the following nodes: ")
		nodes := response.Nodes.Nodes
		contacts := make([]Contact, len(nodes))
		for i := range nodes { // TODO: Make it work for FIND_VALUE
			fmt.Println(nodes[i].KademliaId + " @ " + nodes[i].Address)
			kID := NewKademliaID(nodes[i].KademliaId)
			if !cbc.me.ID.Equals(kID) {
				contacts[i] = NewContact(kID, nodes[i].Address)
			} else {
				fmt.Println("But that's me! I can't add myself now, can I?")
			}
		}
		cbc.onContacts(sender, contacts)
	case *NetworkMessage.ValueResponse_Content:
		fmt.Println("Received the actual data!")
		cbc.onData(sender, response.Content)
	}
}

// FIND_VALUE RPC
func (network *Network) SendFindDataMessage(hash string, recipient *Contact, onData OnDataCallback, onContacts OnContactsCallback, onTimeout func(*KademliaID, *KademliaID)) {
	cbc := &callbackContainer{}
	cbc.onData = onData
	cbc.onContacts = onContacts
	cbc.me = network.Me()
	network.sendFindDataMessage(NewKademliaID(hash), recipient, onTimeout, cbc)
}

func (network *Network) sendFindDataMessage(key *KademliaID, recipient *Contact, onTimeout func(*KademliaID, *KademliaID), cbc *callbackContainer) {
	// TODO
	packet := network.createPacket()
	randomID := network.findTable.MakeRequest(recipient.ID, onTimeout, cbc.handleFindDataResponse)

	packet.FindValue = &NetworkMessage.Find{
		RandomId: randomID.String(),
		Hash:     key.String(),
	}

	out, merr := proto.Marshal(packet)
	if merr != nil {
		fmt.Println("Error marshalling find_node packet")
	}

	network.dw.sendDataToAddress(ensurePort(recipient.Address, "42042"), out)
}

type hashValueContainer struct {
	hash    string
	network *Network
	content []byte
}

func (network *Network) SendStoreIterative(hash string, content []byte) {
	hvc := &hashValueContainer{}
	hvc.network = network
	hvc.hash = hash
	hvc.content = content
	network.NodeLookup(NewKademliaID(hash), hvc.storeOnReceivedContacts)
}

func (hvc *hashValueContainer) storeOnReceivedContacts(contacts []Contact, data *[]byte) {
	for i := range contacts {
		contact := contacts[i]
		if contact.ID != nil {
			hvc.network.SendStoreMessage(hvc.network.CreateStoreMessage(hvc.hash, hvc.content, false), contact.Address)
		}
	}
}

// STORE RPC
func (network *Network) CreateStoreMessage(hash string, content []byte, pin bool) *NetworkMessage.Store {

	randomID := NewRandomKademliaID()

	fmt.Println("Store message created, with hash:" + hash + "With random id: " + randomID.String())
	store := &NetworkMessage.Store{
		RandomId:   randomID.String(),
		KademliaId: network.routingTable.Me().ID.String(),
		Address:    network.routingTable.Me().Address,
		Hash:       hash,
		Content:    content,
		Pin:        pin,
	}
	return store
}

func (network *Network) CreateStoreResponseMessage(randomID *KademliaID) *NetworkMessage.StoreResponse {

	fmt.Println("StoreResponse message created")
	store := &NetworkMessage.StoreResponse{
		RandomId:   randomID.String(),
		KademliaId: network.routingTable.Me().ID.String(),
		Address:    network.routingTable.Me().Address,
	}
	return store
}

func (network *Network) SendStoreMessage(storeMessage *NetworkMessage.Store, address string) {

	packet := network.createPacket()
	packet.Store = storeMessage

	out, merr := proto.Marshal(packet)
	if merr != nil {
		fmt.Println("Error marshaling store packet")
	} else {
		network.rpcTable.Push(NewKademliaID(storeMessage.RandomId))
		fmt.Println("Store adress: " + address)
		network.dw.sendDataToAddress(ensurePort(address, "42042"), out)
	}
}

func (network *Network) SendStoreResponseMessage(storeMessage *NetworkMessage.StoreResponse, address string) {

	packet := network.createPacket()
	packet.StoreResponse = storeMessage

	out, merr := proto.Marshal(packet)
	if merr != nil {
		fmt.Println("Error marshaling storeresponse packet")
	} else {
		//fmt.Println("Marshalled data is " + string(out[:]))
		fmt.Println("Sending StoreResponse with adress: " + address + "With random id: " + storeMessage.RandomId)
		network.dw.sendDataToAddress(ensurePort(address, "42042"), out)
	}
}

func (network *Network) HandleStoreMessage(storeMessage *NetworkMessage.Store) {
	//Recieve data and filename
	randomID := storeMessage.RandomId
	kademliaID := storeMessage.KademliaId
	address := storeMessage.Address
	content := storeMessage.Content
	fileName := storeMessage.Hash
	pin := storeMessage.Pin

	fmt.Println("Recieved store message, frome:" + kademliaID + "With random id: " + randomID)
	contentRes := network.storeTable.Push(content, fileName, true, pin)
	if contentRes == false {
		fmt.Println("ERROR SAVING FILE")
	} else {
		network.SendStoreResponseMessage(network.CreateStoreResponseMessage(NewKademliaID(randomID)), address)
	}

}

func (network *Network) HandleStoreResponseMessage(storeMessage *NetworkMessage.StoreResponse) {
	row := network.rpcTable.Pop(NewKademliaID(storeMessage.RandomId))
	//var contact Contact
	if row == nil {
		fmt.Println("Received StoreResponse with random id " + storeMessage.RandomId + " but nothing was found in the rpcTable")
	}
}
