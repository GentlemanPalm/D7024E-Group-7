package d7024e

import (
	"NetworkMessage"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
	"io/ioutil"
	
	

	"github.com/golang/protobuf/proto"
)

type Network struct {
	routingTable *RoutingTable
	pingTable    *PingTable
}

func NewNetwork(routingTable *RoutingTable) *Network {
	nw := &Network{}
	nw.routingTable = routingTable
	nw.routingTable.Me().Address = getIaddr()
	nw.pingTable = NewPingTable()
	return nw
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
		fmt.Println("Received packet with STORE")
		go network.HandleStoreMessage(packet.Store)
	} else {
		fmt.Println("Received packet, but Store left blank")
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
	network.routingTable.AddContact(contact)
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

func (network *Network) CreateStoreMessage(filePath string) *NetworkMessage.Store {

	//Move this.
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Store message created, with hash:" + filePath)
	store := &NetworkMessage.Store{
		Hash: Hash(filePath),
		Content:    content,
	}
	return store
}

func (network *Network) HandlePingTimeout(randomID *KademliaID, replacement *Contact) {
	time.Sleep(1 * time.Second)
	// Atomic operation, removes the item from the table and returns it
	row := network.pingTable.Pop(randomID)
	// Nil row implies a response was found in time
	if row != nil {
		fmt.Println("PingTable entry for " + randomID.String() + " remained after timeout.")
		if replacement == nil {
			fmt.Println("No replacement was found. Deleting.")
		} else {
			fmt.Println("The replacement has an ID of " + replacement.ID.String())
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
	network.routingTable.AddContact(contact)
	fmt.Println("Got the PONG message for " + pongMessage.KademliaId + " with random ID " + pongMessage.RandomId)

	//Send store when recieving pong, test only.
	network.SendStoreMessage(network.CreateStoreMessage("d7024e/text.txt"), pongMessage.Address)

}

func (network *Network) HandleStoreMessage(storeMessage *NetworkMessage.Store) {
	//Recieve data and filename
	data := storeMessage.Content
	fileName := storeMessage.Hash

	fmt.Println("Recieved store message, with filename:" + fileName)
	
    err := ioutil.WriteFile(fileName, data, 0644)
    if err != nil {
		log.Fatal(err)
	}
	//Test for reading stored file
	content, err2 := ioutil.ReadFile(fileName)
	if err2 != nil {
		log.Fatal(err2)
	}

	fmt.Printf("File contents: %s", content)
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

func createPacket() *NetworkMessage.Packet {
	return &NetworkMessage.Packet{}
}

func ensurePort(address string, port string) string {
	adr := strings.Split(address, ":")
	return adr[0] + ":" + port
}

func (network *Network) SendPongMessage(pongMessage *NetworkMessage.Pong, address string) {
	packet := &NetworkMessage.Packet{}
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

func (network *Network) sendPingPacket(randomID *KademliaID, contact *Contact) {
	packet := createPacket()
	packet.Ping = &NetworkMessage.Ping{
		RandomId:   randomID.String(),
		KademliaId: network.routingTable.Me().ID.String(),
		Address:    network.routingTable.Me().Address,
	}

	out, merr := proto.Marshal(packet)
	if merr != nil {
		fmt.Println("Error marshalling ping packet")
	}

	sendDataToAddress(contact.Address, out)
}

// FIND_NODE RPC
func (network *Network) SendFindContactMessage(contact *Contact) {
	// TODO
	// 1. Send 160-bit key to recipient
	// 2. Expect k number of triplets on the form of <ip, port, nodeid>
	//    (may get less than that if the recipient node does not know of k nodes)
	// 3. Return this information so it can be processed elsewhere
	//    (to send STORE RPCs or just update the routing table?)
}

// FIND_VALUE RPC
func (network *Network) SendFindDataMessage(hash string) {
	// TODO
}

// STORE RPC
func (network *Network) SendStoreMessage(storeMessage *NetworkMessage.Store, address string) {
	fmt.Println("SEND MESSAGE STORE")
	packet := createPacket()
	packet.Store = storeMessage

	
	out, merr := proto.Marshal(packet)
	if merr != nil {
		fmt.Println("Error marshaling store packet")
	} else {
		//fmt.Println("Marshalled data is " + string(out[:]))
		fmt.Println("Store adress: " + address)
		sendDataToAddress(ensurePort(address, "42042"), out)
	}
}



