package d7024e

import (
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

func (server *Server) createResponse(response *Response) string {
	marsh, merr := json.Marshal(response)

	if merr != nil {
		fmt.Println(merr)
	}

	s := string(marsh[:len(marsh)])

	return s
}

// Based around the examples detailed in https://golang.org/doc/articles/wiki/
func (server *Server) cat(w http.ResponseWriter, r *http.Request) {
	fmt.Println("received HTTP request!")
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func (server *Server) store(w http.ResponseWriter, r *http.Request) {
	fmt.Println("received HTTP store request")
	request := parseRequest(r)

	//storeTable := network.GetStoreTable()
	//storeTable.Push(content, fileName, true, true)

	fmt.Fprintf(w, "Snark snark %s!", r.URL.Path[1:])
}

func (server *Server) pin(w http.ResponseWriter, r *http.Request) {
	fmt.Println("received pin HTTP request!")

	fmt.Fprintf(w, "abcd")
}

func (server *Server) unpin(w http.ResponseWriter, r *http.Request) {
	fmt.Println("received snarky HTTP request!")
	fmt.Fprintf(w, "Woof woof %s!", r.URL.Path[1:])
}

func StartServer(network *Network) {
	fmt.Println("Launched web server!")
	server := &Server{}
	server.network = network
	http.HandleFunc("/pin/", server.pin)
	http.HandleFunc("/unpin/", server.unpin)
	http.HandleFunc("/store/", server.store)
	http.HandleFunc("/cat/", server.cat)
	http.ListenAndServe(":8080", nil)
}
