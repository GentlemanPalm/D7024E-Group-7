package d7024e

// This file intends to implement a table of ping calls and their timeouts

import (
	"sync"
	//"d7024e/kademliaid"
)

// The following data is saved for each row in the ping table:
// 1. Unique ping identifier. This is a random kadmelia ID
// 2. The KademliaID of the item in question. One can derive the bucket ID from this

// A ping should make an entry in this table and THEN send the ping over the network
// Then a timeout goroutine should be launched.

// If a response is received from the PONG, then the table row should be deleted
// and the bucket updated with the response data

type row struct {
	randomID   *KademliaID // The rand
	kademliaID *KademliaID //
}

type PingTable struct {
	rows []row
	lock *sync.Mutex
}

func NewPingTable() *PingTable {
	table := &PingTable{}
	table.rows = make([]row, 2*20) // TODO: Make global variable for K
	table.lock = &sync.Mutex{}
	return table
}

func (table *PingTable) Push(randomID *KademliaID, kademliaID *KademliaID) {
	table.lock.Lock()
	defer table.lock.Unlock()
	table.rows = append(table.rows, row{randomID, kademliaID})
}

// Get and remove a row with the given id
// Returns nil if the block wasn't found
// Untested and unlikely to work as intended
func (table *PingTable) Pop(id *KademliaID) *row {
	table.lock.Lock()
	defer table.lock.Unlock()
	for i := 0; i < len(table.rows); i = i + 1 {
		if table.rows[i].randomID == nil {
			continue
		}
		if table.rows[i].randomID.Equals(id) {
			item := table.rows[i]
			table.rows = table.rows[:i+copy(table.rows[i:], table.rows[i+1:])]
			return &item
		}
	}
	return nil
}
