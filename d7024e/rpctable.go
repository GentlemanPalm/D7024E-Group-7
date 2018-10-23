package d7024e

// This file intends to implement a table of ping calls and their timeouts

import (
	"sync"
	//"d7024e/kademliaid"
)


type field struct {
	randomID   *KademliaID // The rand //
}

type RpcTable struct {
	rows map[string]*field
	lock *sync.Mutex
}

func NewRpcTable() *RpcTable {
	table := &RpcTable{}
	table.rows = make(map[string]*field)
	table.lock = &sync.Mutex{}
	return table
}

func (table *RpcTable) Push(randomID *KademliaID) {
	rpc := &field{randomID}
	table.lock.Lock()
	defer table.lock.Unlock()
	table.rows[randomID.String()] = rpc
	
}

// Get and remove a row with the given id
// Returns nil if the block wasn't found
// Untested and unlikely to work as intended
func (table *RpcTable) Pop(randomId *KademliaID) *KademliaID {

	table.lock.Lock()
	defer table.lock.Unlock()
	if len(table.rows) < 1 {
		return nil
	}else{
		item := table.rows[randomId.String()].randomID
		delete(table.rows, randomId.String())
		return item
	}
	
}