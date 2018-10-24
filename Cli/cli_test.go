package Cli

import (
	"testing"
	"fmt"
)

/*func mockS(sentPacket *PacketContainer) *Server {
	network := mockNetwork()

	mfh := &mockFileHandler{}
	mfh.kvt = make(map[string][]byte)
	network.storeTable.fh = mfh

	mdw := &MockDataWriter{}
	mdw.callback = sentPacket.findNodeCallback
	network.dw = mdw
	server := &Server{}
	server.network = network
	return server
}*/

func TestCli(t *testing.T) {
	
	//server := mockS()
	hash := Hash("Hello")
	if hash == nil {
		t.Error("HAsh fail")
	}else{
		fmt.Println("------------------------------------------------------------------------------------------------------")
		fmt.Println("------------------------------------------------------------------------------------------------------")
		fmt.Println("------------------------------------------------------------------------------------------------------")
		fmt.Println("------------------------------------------------------------------------------------------------------")
		fmt.Println(hash)
	} 



}
