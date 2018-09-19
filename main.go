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
)

const defaultPort = "42042"

func main() {
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

	//Test for hash
	var str = "d7024e/text.txt"
	fmt.Println("\n" + "Hello, I hashed this file: ")
	d7024e.Hash(str)

	ioutil.WriteFile("come_on", []byte("asdf"), 0644)

	go listenForConnections()

	//simple Read
	//buffer := make([]byte, 1024)
	//conn.Read(buffer)

	for {
		//fmt.Println("hue")
	}
}

func sendStuff() {
	//Connect udp
	saddr, e0 := net.ResolveUDPAddr("udp", "kademliaBootstrap:"+defaultPort)
	if e0 != nil {
		fmt.Println(e0)
	}
	conn, err := net.DialUDP("udp", nil, saddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	rnum := rand.Int() % 100000
	//simple write
	fmt.Println("Trying to write a packet of data with " + strconv.Itoa(rnum))
	conn.Write([]byte(strconv.Itoa(rnum)))
	fmt.Println("Wrote a packet of data")
}

func replyTo(uaddr *net.UDPAddr) {
	conn, err := net.DialUDP("udp", nil, uaddr)
	defer conn.Close()

	_, err2 := conn.Write([]byte("Fuck you"))
	if err != nil || err2 != nil {
		fmt.Println("Error when replying to the message")
		fmt.Println(err)
		fmt.Println(err2)
	}
	fmt.Println("Replied to " + uaddr.String() + " and told it to fuck off")
}

func listenForConnections() {
	// Taken almost directly from
	// http://www.minaandrawos.com/2016/05/14/udp-vs-tcp-in-golang/

	// listen to incoming udp packets
	var nrofPacketsRcvd int = 0
	fmt.Println("Entering listenForConnections")

	saddr, _ := net.ResolveUDPAddr("udp", ":"+defaultPort)

	pc, err := net.ListenUDP("udp", saddr)
	if err != nil {
		log.Fatal(err)
	}
	defer pc.Close()
	fmt.Println("Now listening for connections")

	go sendStuff()

	for {
		//simple read
		buffer := make([]byte, 1024)
		fmt.Print("Reading from ListenPacket...")
		_, addr, err := pc.ReadFromUDP(buffer)

		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(" done")

		nrofPacketsRcvd = nrofPacketsRcvd + 1
		s := string(buffer[:14])

		fmt.Println("Received: " + s + " from " + addr.String())

		go replyTo(addr)

		ioutil.WriteFile("packetsreceived", []byte(strconv.Itoa(nrofPacketsRcvd)), 0644)

		//simple write
		//pc.WriteTo([]byte("Hello from client"), net.ResolveUDPAddr("udp", ":2000"))
	}
}
