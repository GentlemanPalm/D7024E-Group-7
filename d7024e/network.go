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
	routingTable *RoutingTable
	pingTable    *PingTable
	findTable    *FindTable
}

func NewNetwork(routingTable *RoutingTable) *Network {
	nw := &Network{}
	nw.routingTable = routingTable
	nw.routingTable.Me().Address = getIaddr()
	nw.pingTable = NewPingTable()
	nw.findTable = NewFindTable()
	return nw
}

func (network *Network) Me() *Contact {
	return network.routingTable.Me()
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

func Listen(ip string, port int) {
	// TODO
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
		network.HandleOriginMessage(packet.Origin)
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

}

func (network *Network) HandleOriginMessage(origin *NetworkMessage.KademliaPair) {

	fmt.Println("Received an origin message")
	fmt.Println("id=" + origin.KademliaId + " addr=" + origin.Address)
	if !network.Me().ID.Equals(NewKademliaID(origin.KademliaId)) {
		fmt.Println("Added contact")
		network.routingTable.AddContact(NewContact(NewKademliaID(origin.KademliaId), origin.Address))
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
		network.routingTable.AddContact(contact)
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

func (network *Network) HandlePingTimeout(randomID *KademliaID, replacement *Contact) {
	time.Sleep(time.Duration(1) * time.Second)
	// Atomic operation, removes the item from the table and returns it
	row := network.pingTable.Pop(randomID)
	// Nil row implies a response was found in time
	if row != nil {
		fmt.Println("PingTable entry for " + randomID.String() + " remained after timeout.")
		if replacement == nil {
			fmt.Println("No replacement was found. Deleting.")
		} else {
			fmt.Println("The replacement has an ID of " + replacement.ID.String() + " TODO: IMPLEMENT")
		}
	} else {
		fmt.Println("Looked to the table for " + randomID.String() + " but received a response in time")
	}
}

func (network *Network) HandlePongMessage(pongMessage *NetworkMessage.Pong) {
	// Atomically remove the item from the table and get the row
	row := network.pingTable.Pop(NewKademliaID(pongMessage.KademliaId))
	var contact Contact
	if row == nil {
		fmt.Println("Received pong with random id " + pongMessage.RandomId + " but nothing was found in the ping table")
		contact = NewContact(NewKademliaID(pongMessage.KademliaId), pongMessage.Address)
	} else {
		contact = NewContact(row.kademliaID, pongMessage.Address)
	}
	// Does this simply work??
	if !network.Me().ID.Equals(contact.ID) {
		network.routingTable.AddContact(contact)
	} else {
		fmt.Println("Recevied pong from self. Not adding to contact list")
	}
	fmt.Println("Got the PONG message for " + pongMessage.KademliaId + " with random ID " + pongMessage.RandomId)
}

func sendDataToAddress(address string, data []byte) {
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
		sendDataToAddress(ensurePort(address, "42042"), out)
	}

}

// PING RPC
func (network *Network) SendPingMessage(contact *Contact) {
	network.SendPingMessageWithReplacement(contact, nil)
}

func (network *Network) SendPingMessageWithReplacement(contact *Contact, replacement *Contact) {
	randomID := NewRandomKademliaID()
	network.pingTable.Push(randomID, contact.ID)
	go network.HandlePingTimeout(randomID, replacement)
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

	sendDataToAddress(contact.Address, out)
}

// FIND_NODE RPC
func (network *Network) SendFindContactMessage(key *KademliaID, recipient *Contact) {
	// TODO
	packet := network.createPacket()
	randomID := network.findTable.MakeRequest(recipient.ID, nil, network.handleFindContactResponse)

	packet.FindNode = &NetworkMessage.Find{
		RandomId: randomID.String(),
		Hash:     key.String(),
	}

	out, merr := proto.Marshal(packet)
	if merr != nil {
		fmt.Println("Error marshalling find_node packet")
	}

	sendDataToAddress(ensurePort(recipient.Address, "42042"), out)
	// 1. Send 160-bit key to recipient
	// 2. Expect k number of triplets on the form of <ip, port, nodeid>
	//    (may get less than that if the recipient node does not know of k nodes)
	// 3. Return this information so it can be processed elsewhere
	//    (to send STORE RPCs or just update the routing table?)
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
				network.routingTable.AddContact(NewContact(kID, nodes[i].Address))
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
	// TODO: Add contact to buckets?
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

	sendDataToAddress(ensurePort(addr, "42042"), out)
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

func (network *Network) NodeLookup(id *KademliaID) {
	// TODO
	// 1. Create a shortlist with the three closest items of the id
	//     determine which element is the closest element
	//
	// 2. Send FIND_NODE or FIND_VALUE requests to the three nodes
	//     mark the items as visited in the table.
	//
	// 3. When the results of the requests arrive, update the table with
	//     the results. Determine the new closest elements. Kick out
	//      elements beyond the 'k' (20) closest elements.
	//
	// 4. For each request returning or timing out, pick off the next nearest
	//     element, mark it as 'visited' and send new FIND_* request to it.
	//      Mark elements which timed out as 'dead'. Keep dead nodes in sep table?
}

// FIND_VALUE RPC
func (network *Network) SendFindDataMessage(hash string) {
	// TODO
}

// STORE RPC
func (network *Network) SendStoreMessage(data []byte) {
	// TODO
}
