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

	buffer := make([]byte, 8192)

	for {
		//simple read

		fmt.Print("Reading from ListenPacket...")
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
	if packet == nil {

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
	network.pingTable.Pop(NewKademliaID(pongMessage.KademliaId))
	fmt.Println("Got the PONG message for " + pongMessage.KademliaId + " with random ID " + pongMessage.RandomId + " TODO: Implement update of pong packet")
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
	// TODO
	// 1. Send a PING Message to the contact
	// 2. Set up a listener for the PONG reply
	// 3. Update routing table for when PONG is received
	// 4. Allow this RPC to be piggybacked onto other RPCs?
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

// FIND_NODE RPC
func (network *Network) SendStoreMessage(data []byte) {
	// TODO
}
