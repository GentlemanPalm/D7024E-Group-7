package d7024e

import (
	"NetworkMessage"
	"fmt"
	"sync"
	"time"
)

// This table intends to keep track of which FIND_* requests have been sent,
// so that one can hook a callback for when the data is available.

type frow struct {
	randomID   *KademliaID
	recipient  *KademliaID
	onTimeout  func(*KademliaID, *KademliaID)
	onResponse func(*KademliaID, *NetworkMessage.ValueResponse)
}

type FindTable struct {
	rows []frow
	lock *sync.Mutex
}

func NewFindTable() *FindTable {
	table := &FindTable{}
	table.lock = &sync.Mutex{}
	table.rows = make([]frow, 2*20)
	return table
}

func (ft *FindTable) ProcessResult(response *NetworkMessage.ValueResponse) {
	ft.lock.Lock()
	defer ft.lock.Unlock()

	elem := ft.deleteRandomKey(NewKademliaID(response.RandomId))
	if elem != nil {
		fmt.Println("Received a response from " + elem.recipient.String() + " regarding " + elem.randomID.String())
		if elem.onResponse != nil {
			go elem.onResponse(elem.recipient, response)
		} else {
			fmt.Println("Got the response, but no response trigger was set")
		}
	} else {
		fmt.Println("Received a response from " + response.RandomId + ", but nothing was found. Timeout?")
	}

}

func (ft *FindTable) MakeRequest(recipient *KademliaID,
	onTimeout func(*KademliaID, *KademliaID),
	onResponse func(*KademliaID, *NetworkMessage.ValueResponse)) *KademliaID {

	ft.lock.Lock()
	defer ft.lock.Unlock()

	randomID := NewRandomKademliaID()

	ft.rows = append(ft.rows, frow{randomID, recipient, onTimeout, onResponse})

	go ft.timeout(onTimeout, randomID)

	return randomID

}

func (table *FindTable) deleteRandomKey(key *KademliaID) *frow {
	for i := 0; i < len(table.rows); i = i + 1 {
		if table.rows[i].randomID == nil {
			continue
		}
		if table.rows[i].randomID.Equals(key) {
			item := table.rows[i]
			table.rows = table.rows[:i+copy(table.rows[i:], table.rows[i+1:])]
			return &item
		}
	}
	return nil
}

func (ft *FindTable) timeout(onTimeout func(*KademliaID, *KademliaID), rID *KademliaID) {
	time.Sleep(time.Duration(1) * time.Second)

	ft.lock.Lock()
	defer ft.lock.Unlock()

	elem := ft.deleteRandomKey(rID)

	if elem != nil {
		fmt.Println("Timeout occurred and " + rID.String() + " was discarded.")
		if onTimeout != nil {
			onTimeout(elem.recipient, elem.randomID)
		} else {
			fmt.Println("Timeout occurred, but the onTimeout function wasn't set")
		}
	} else {
		fmt.Println("Timeout occurred, but response has already been processed for " + rID.String())
	}
}
