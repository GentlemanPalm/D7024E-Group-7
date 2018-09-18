package d7024e

type Network struct {
}

func Listen(ip string, port int) {
	// TODO
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
