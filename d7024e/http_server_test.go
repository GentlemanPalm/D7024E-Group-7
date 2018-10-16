package d7024e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

const hello_world_b64 = "SGVsbG8sIFdvcmxk" // "Hello, World" in base64

func mockServer(sentPacket *PacketContainer) *Server {
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
}

func sendRequest(server *Server, endpoint string, request *Request) *Response {
	buffer := server.marshalRequest(request)

	bts := []byte(buffer)

	rq, err := http.NewRequest("POST", "http://localhost:8080/"+endpoint+"/", bytes.NewBuffer(bts))

	client := &http.Client{}

	resp, err := client.Do(rq)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	data := &Response{}
	result := json.Unmarshal(body, data)
	if result != nil {
		fmt.Println(result)
	}
	return data
}

func TestStoreRequest(t *testing.T) {
	pc := &PacketContainer{}
	server := mockServer(pc)
	request := &Request{}
	hash := NewRandomKademliaID()
	request.Hash = hash.String()
	request.Content = hello_world_b64 // "Hello, World" in base64

	response := sendRequest(server, "store", request)
	if response.Status != "ok" && response.Content != request.Hash {
		t.Error("Expected 'ok' status and content to equal the request hash")
	}
}

func TestCatRequest(t *testing.T) {
	pc := &PacketContainer{}
	server := mockServer(pc)
	request := &Request{}
	hash := NewRandomKademliaID()
	request.Hash = hash.String()
	content := hello_world_b64
	request.Content = content

	response := sendRequest(server, "store", request)
	if response.Status != "ok" && response.Content != request.Hash {
		t.Error("Expected 'ok' status and content to equal the request hash")
	}

	request.Content = ""
	response = sendRequest(server, "cat", request)
	if response.Status != "ok" && response.Content != content {
		t.Error("The content returned by the response is not what was sent to the server")
	}

	hash = NewRandomKademliaID()
	request.Hash = hash.String()
	response = sendRequest(server, "cat", request)
	if response.Status == "ok" {
		t.Error("Somehow some random ID was found in the system. This should never happen")
	}

}

func TestPinRequest(t *testing.T) {
	pc := &PacketContainer{}
	server := mockServer(pc)
	request := &Request{}
	hash := NewRandomKademliaID()
	request.Hash = hash.String()
	content := hello_world_b64
	request.Content = content

	response := sendRequest(server, "pin", request)
	if response.Status == "ok" {
		t.Error("Should not receive an OK response when pinning a non-existant item")
	}

	response = sendRequest(server, "store", request)
	if response.Status != "ok" {
		t.Error("Expected to be able to store item for pinning")
	}

	response = sendRequest(server, "pin", request)
	if response.Status != "ok" {
		t.Error("Expected to be able to pin existing item")
	}
}

func TestUnpinRequest(t *testing.T) {
	pc := &PacketContainer{}
	server := mockServer(pc)
	request := &Request{}
	hash := NewRandomKademliaID()
	request.Hash = hash.String()
	content := hello_world_b64
	request.Content = content

	response := sendRequest(server, "unpin", request)
	if response.Status == "ok" {
		t.Error("Should not receive an OK response when unpinning a non-existant item")
	}

	response = sendRequest(server, "store", request)
	if response.Status != "ok" {
		t.Error("Expected to be able to store item for pinning")
	}

	response = sendRequest(server, "unpin", request)
	if response.Status != "ok" {
		t.Error("Expected to be able to unpin existing (and pinned) item")
	}
}
