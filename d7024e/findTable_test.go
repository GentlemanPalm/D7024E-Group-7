package d7024e

import (
	"NetworkMessage"
	"fmt"
	"testing"
	"time"
)

// MakeRequest(recipient *KademliaID,
//	onTimeout func(*KademliaID, *KademliaID),
//	onResponse func(*KademliaID, *NetworkMessage.ValueResponse)) *KademliaID

type t2 struct {
	t   *testing.T
	id  *KademliaID
	rid *KademliaID
}

func (t *t2) testOnTimeoutErr(sender *KademliaID, randomID *KademliaID) {
	t.t.Error("Error: Received timeout when expecting response")
}

func (t *t2) testOnTimeout(sender *KademliaID, randomID *KademliaID) {
	if sender.String() != t.id.String() {
		t.t.Error("Sender unknown for the timeout")
	} else {
		t.t.Log("Hit timeout as expected")
	}
}

func (t *t2) testOnResponse(sender *KademliaID, response *NetworkMessage.ValueResponse) {
	if response.RandomId != t.rid.String() {
		t.t.Error("Error: Expected RandomId for pushed response")
	} else {
		t.t.Log("Received response as expected")
	}
}

func TestFindTable(t *testing.T) {
	ft := NewFindTable()
	id := NewRandomKademliaID()
	t1 := &t2{}
	t1.t = t
	t1.id = NewRandomKademliaID()
	t1.rid = NewRandomKademliaID()
	randomId := ft.MakeRequest(id, t1.testOnTimeoutErr, t1.testOnResponse)
	response := &NetworkMessage.ValueResponse{}
	response.RandomId = randomId.String()
	t1.rid = randomId
	fmt.Println(response.RandomId)
	ft.ProcessResult(response)
	ft.MakeRequest(t1.id, t1.testOnTimeout, t1.testOnResponse)
	time.Sleep(time.Duration(2) * time.Second)
}
