package d7024e

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Server struct {
	network *Network
}

// Data structure used by all requests
type Request struct {
	Hash    string `json:"hash"`
	Content string `json"content,omitempty"`
}

// The generic response used for all requests
type Response struct {
	Status  string `json:"status"`
	Content string `json:"content,omitempty"`
}

// Get the JSON request from the message body. TODO: Make better error checking
func (server *Server) parseRequest(r *http.Request) *Request {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	data := &Request{}
	result := json.Unmarshal(body, data)
	if result != nil {
		fmt.Println(result)
	}
	return data
}

func (server *Server) marshalResponse(response *Response) string {
	marsh, merr := json.Marshal(response)

	if merr != nil {
		fmt.Println(merr)
	}

	s := string(marsh[:len(marsh)])

	return s
}

func (server *Server) marshalRequest(request *Request) string {
	marsh, merr := json.Marshal(request)

	if merr != nil {
		fmt.Println(merr)
	}

	s := string(marsh[:len(marsh)])

	return s
}

type HttpCallbackContainer struct {
	server *Server
	r      *Request
	w      *http.ResponseWriter
	c      chan string
}

// Based around the examples detailed in https://golang.org/doc/articles/wiki/
func (server *Server) cat(w http.ResponseWriter, r *http.Request) {
	fmt.Println("meow")

	hcc := &HttpCallbackContainer{}
	hcc.server = server
	hcc.r = server.parseRequest(r)
	hcc.w = &w
	hcc.c = make(chan string)
	content := server.network.storeTable.Get(hcc.r.Hash)
	if content != nil { // If stored locally
		response := &Response{}
		response.Status = "ok"
		b64value := base64.StdEncoding.EncodeToString(content)
		response.Content = b64value
		result := server.marshalResponse(response)
		fmt.Println("[CAT CACHED] " + result)
		fmt.Fprintf(*hcc.w, result)
		return
	}
	go server.network.ValueLookup(NewKademliaID(hcc.r.Hash), hcc.onCatCallback)
	result := <-hcc.c
	fmt.Println("[CAT DOEN] " + result)
	fmt.Fprintf(*hcc.w, result)
}

func (hcc *HttpCallbackContainer) onCatCallback(contacts []Contact, content *[]byte) {
	if content == nil {
		fmt.Println("[CAT :3] Could not find value for " + hcc.r.Hash)
		response := &Response{}
		response.Status = "not found"
		response.Content = ""
		hcc.c <- hcc.server.marshalResponse(response)
		return
	}
	value := base64.StdEncoding.EncodeToString(*content)
	response := &Response{}
	response.Status = "ok"
	response.Content = value
	fmt.Println("[CAT :3] Returning value for " + hcc.r.Hash)
	hcc.c <- hcc.server.marshalResponse(response)
}

func (server *Server) store(w http.ResponseWriter, r *http.Request) {
	fmt.Println("received HTTP store request")
	request := server.parseRequest(r)

	storeTable := server.network.GetStoreTable()

	res, err := base64.StdEncoding.DecodeString(request.Content)
	if err != nil {
		fmt.Println(err)
	}
	storeTable.Push(res, request.Hash, true, true)

	response := &Response{}
	response.Content = request.Hash
	response.Status = "ok"

	fmt.Fprintf(w, server.marshalResponse(response))
}

func (server *Server) pin(w http.ResponseWriter, r *http.Request) {
	fmt.Println("received pin HTTP request!")

	request := server.parseRequest(r)

	storeTable := server.network.GetStoreTable()
	result := storeTable.Pin(request.Hash)
	status := ""

	if result {
		status = "ok"
		response := &Response{}
		response.Content = ""
		response.Status = status
		fmt.Fprintf(w, server.marshalResponse(response))
		return
	} else {
		fmt.Println("Received pin, but content not stored on node.")
		status = "not stored on node"
		hcc := &HttpCallbackContainer{}
		hcc.server = server
		hcc.r = request
		hcc.w = &w
		hcc.c = make(chan string)
		go server.network.ValueLookup(NewKademliaID(hcc.r.Hash), hcc.onPinCallback)
		result := <-hcc.c
		fmt.Println("[PIN DONE] " + result)
		fmt.Fprintf(*hcc.w, result)
	}

}

func (hcc *HttpCallbackContainer) onPinCallback(contacts []Contact, content *[]byte) {
	if content == nil {
		fmt.Println("[PIN FAIL] Could not find value for " + hcc.r.Hash)
		response := &Response{}
		response.Status = "not found"
		response.Content = ""
		hcc.c <- hcc.server.marshalResponse(response)
		return
	}
	value := base64.StdEncoding.EncodeToString(*content)
	response := &Response{}
	response.Status = "ok"
	response.Content = value
	hcc.server.network.storeTable.Push(*content, hcc.r.Hash, true, true)
	fmt.Println("[PIN SUCCESS] " + hcc.r.Hash + " is pinned. Returning response.")
	hcc.c <- hcc.server.marshalResponse(response)
}

func (server *Server) unpin(w http.ResponseWriter, r *http.Request) {
	fmt.Println("received unpin HTTP request!")

	request := server.parseRequest(r)

	storeTable := server.network.GetStoreTable()
	result := storeTable.Unpin(request.Hash)
	status := ""

	if result {
		status = "ok"
	} else {
		status = "not found" // TODO: Fetch the item from the network and save it
	}

	response := &Response{}
	response.Content = ""
	response.Status = status

	fmt.Fprintf(w, server.marshalResponse(response))
}

func startServer(server *Server) {
	http.HandleFunc("/pin/", server.pin)
	http.HandleFunc("/unpin/", server.unpin)
	http.HandleFunc("/store/", server.store)
	http.HandleFunc("/cat/", server.cat)
	http.ListenAndServe(":8080", nil)
}

func StartServer(network *Network) {
	fmt.Println("Launched web server!")
	server := &Server{}
	server.network = network
	startServer(server)
}
