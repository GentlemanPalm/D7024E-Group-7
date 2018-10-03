package main

import (
	"NetworkMessage"
	"d7024e"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
)

const defaultPort = "42042"

func main() {
	//Get globals
	globals := d7024e.GetGlobals()
	fmt.Println(globals)

	// This section intends to parse command line parameters
	var port = flag.String("port", defaultPort, "Port to expect connections to")
	var bootstrapIP = flag.String("bsip", "kademliaBootstrap", "IP or network hostname of bootstrap node")
	var bootstrapPort = flag.String("bsport", defaultPort, "Destination port of bootstrap node")

	flag.Parse()

	fmt.Println("Bootstrapping IP and Port: " + *bootstrapIP + ":" + *bootstrapPort)
	fmt.Println("Local port is " + *port)

	var x *d7024e.KademliaID
	x = d7024e.NewRandomKademliaID()
	var y = NetworkMessage.SearchRequest{}
	fmt.Println("Hello, World! " + x.String() + y.String())

	rand.Seed(int64(time.Now().Nanosecond()))
	me := d7024e.NewContact(d7024e.NewRandomKademliaID(), getIaddr())
	fmt.Println("Created contact for myself")
	fmt.Println("Reachable as " + me.ID.String() + " " + me.Address)
	routingTable := d7024e.NewRoutingTable(me)
	network := d7024e.NewNetwork(routingTable)

	sport, _ := strconv.Atoi(*port)
	go send2(me.ID, network)

	go testFindNode(me.ID, network, 14)
	//go testFindNode(me.ID, network, 25)
	go testNodeLookup(me.ID, network, 20)
	//go testValueLookup(me.ID, network, 25)

	go marcusTest(network)

	network.Listen(sport)
	//go listenForConnections()

	//simple Read
	//buffer := make([]byte, 1024)
	//conn.Read(buffer)

	for {
		//fmt.Println("hue")
	}
}

func marcusTest(network *Network) {
	time.Sleep(time.Duration(1) * time.Second)
	//Bootsrap gets pinned file for republishing.
	saddr, e0 := net.ResolveUDPAddr("udp", "kademliaBootstrap:42042")
	if e0 != nil {
		fmt.Println(e0)
	}
	strRemoteAddr := saddr.String()
	ip := getIaddr() + ":42042"
	if ip == strRemoteAddr {
		fileName := d7024e.Hash("d7024e/text.txt")
		filePath := "Files/" + fileName

		content, err1 := ioutil.ReadFile("d7024e/text.txt")
		if err1 != nil {
			log.Fatal(err1)
		}

		err2 := ioutil.WriteFile(filePath, content, 0644)
		if err2 != nil {
			log.Fatal(err2)
		} else {

			storeTable := network.GetStoreTable()
			storeTable.Push(content, fileName, true, true)
		}
	}

	go testRepublish(me.ID, network)
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

func testRepublish(kademliaId *d7024e.KademliaID, network *d7024e.Network) {
	network.Republish(kademliaId)
}

func testNodeLookup(id *d7024e.KademliaID, network *d7024e.Network, delay int) {
	time.Sleep(time.Duration(delay) * time.Second)
	fmt.Println("At last, NodeLookup. Not just FindNode.")
	network.NodeLookup(id, printContacts)
}

func testValueLookup(id *d7024e.KademliaID, network *d7024e.Network, delay int) {
	time.Sleep(time.Duration(delay) * time.Second)
	fmt.Println("--> WOOO!!! Value Lookup!!! <<--")
	network.ValueLookup(id, printContacts)
}

func printContacts(contacts []d7024e.Contact, notUsed *[]byte) {
	fmt.Println("Holy crap. NodeLookup finished!")
	for i := range contacts {
		if contacts[i].ID != nil {
			fmt.Println(contacts[i].ID.String() + " @@@ " + contacts[i].Address)
		} else {
			fmt.Println("Had a NIL value for one of contacts")
		}
	}
	if notUsed != nil {
		fmt.Println("Actually Received a value!!!!")
	}
}

func testFindNode(target *d7024e.KademliaID, network *d7024e.Network, delay int) {
	time.Sleep(time.Duration(delay) * time.Second)
	fmt.Println("Waited the timeout period. Now doing node lookup")
	network.SendFindContactMessage(d7024e.NewRandomKademliaID(), network.Me())
}

func send2(kademliaId *d7024e.KademliaID, network *d7024e.Network) {
	fmt.Println("Entered send 2")
	//Connect udp
	time.Sleep(time.Duration(9) * time.Second)
	fmt.Println("Attempting to send initial ping...")
	time.Sleep(time.Duration(1) * time.Second)

	saddr, e0 := net.ResolveUDPAddr("udp", "kademliaBootstrap:42042")
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
	fmt.Println("Trying to write a packet of data")
	ping := &NetworkMessage.Ping{
		RandomId:   d7024e.NewRandomKademliaID().String(),
		KademliaId: kademliaId.String(),
		Address:    getIaddr(),
	}

	packet := &NetworkMessage.Packet{}
	packet.Ping = ping

	out, merr := proto.Marshal(packet)
	if merr != nil {
		fmt.Println("Error marshaling ping packet")
	} else {
		//fmt.Println("Marshalled data is " + string(out[:]))
		fmt.Println("RandomId is " + ping.RandomId)
	}
	_, werr := conn.Write(out)
	if werr != nil {
		fmt.Println("Something went wrong with sending inital packet")
		fmt.Println(werr)
	}
	fmt.Println("Wrote a packet of data")
}

func sendStuff() {
	//Connect udp
	time.Sleep(time.Duration(10) * time.Second)
	saddr, e0 := net.ResolveUDPAddr("udp", "kademliaBootstrap:42042")
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
	fmt.Println("Trying to write a packet of data")
	ping := &NetworkMessage.Ping{
		RandomId:   d7024e.NewRandomKademliaID().String(),
		KademliaId: "AAAAAAAAAAAAAAAAAAAA",
		Address:    getIaddr(),
	}

	out, merr := proto.Marshal(ping)
	if merr != nil {
		fmt.Println("Error marshaling ping packet")
	} else {
		//fmt.Println("Marshalled data is " + string(out[:]))
		fmt.Println("RandomId is " + ping.RandomId)
	}
	_, werr := conn.Write(out)
	if werr != nil {
		fmt.Println("Something went wrong with sending inital packet")
		fmt.Println(werr)
	}
	fmt.Println("Wrote a packet of data")
}

func replyTo(uaddr net.Addr) {
	adr := strings.Split(uaddr.String(), ":")
	conn, err := net.Dial("udp", adr[0]+":"+defaultPort)
	defer conn.Close()

	ping := &NetworkMessage.Ping{
		RandomId:   d7024e.NewRandomKademliaID().String(),
		KademliaId: "AAAAAAAAAAAAAAA",
		Address:    "0.0.0.0",
	}

	out, merr := proto.Marshal(ping)
	if merr != nil {
		fmt.Println("Error marshaling ping packet")
	} else {
		//fmt.Println("Marshalled data is " + string(out[:]))
		fmt.Println("RandomId is " + ping.RandomId)
	}

	_, err2 := conn.Write(out)
	if err != nil || err2 != nil {
		fmt.Println("Error when replying to the message")
		fmt.Println(err)
		fmt.Println(err2)
	}
	fmt.Println("Replied to " + adr[0] + ":" + defaultPort + " and told it to fuck off")
}

func listenForConnections() {
	// Taken almost directly from
	// http://www.minaandrawos.com/2016/05/14/udp-vs-tcp-in-golang/

	// listen to incoming udp packets
	//var nrofPacketsRcvd int = 0
	fmt.Println("Entering listenForConnections")

	//saddr, _ := net.ResolveUDPAddr("udp", ":"+defaultPort)

	pc, err := net.ListenPacket("udp", ":"+defaultPort)
	if err != nil {
		log.Fatal(err)
	}
	defer pc.Close()
	fmt.Println("Now listening for connections")

	go sendStuff()

	buffer := make([]byte, 4096)

	for {
		//simple read

		fmt.Print("Reading from ListenPacket...")
		size, addr, err := pc.ReadFrom(buffer)
		go doAsync(buffer, addr, err, size)

		//simple write
		//pc.WriteTo([]byte("Hello from client"), net.ResolveUDPAddr("udp", ":2000"))
	}
}

func doAsync(buffer []byte, addr net.Addr, err error, size int) {
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(" done")

	ping := &NetworkMessage.Ping{}
	marshalerr := proto.Unmarshal(buffer[:size], ping)
	if marshalerr != nil {
		fmt.Println("Received an error from the ping command")
		fmt.Println(marshalerr)
	}
	//	s := string(buffer[:14])

	fmt.Println("Received: " + ping.RandomId + " from " + addr.String())

	go replyTo(addr)

	//ioutil.WriteFile("packetsreceived", []byte(strconv.Itoa(nrofPacketsRcvd)), 0644)
}
