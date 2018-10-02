package d7024e

// This file intends to implement a table of ping calls and their timeouts

import (
	"sync"
	"fmt"
	//"d7024e/kademliaid"
)


type storepath struct {
	randomID   *KademliaID
	hash string 
	path string 
}

type StoreTable struct {
	rows map[string]*storepath
	lock *sync.Mutex
}

func NewStoreTable() *StoreTable {
	table := &StoreTable{}
	table.rows = make(map[string]*storepath)
	table.lock = &sync.Mutex{}
	return table
}

func (table *StoreTable) Push(randomID *KademliaID, hash string, path string) {
	rpc := &storepath{randomID, hash , path}
	table.lock.Lock()
	defer table.lock.Unlock()
	table.rows[randomID.String()] = rpc
	//fmt.Println("VALUES AT INDEX : ")
	//fmt.Println(table.rows[randomID.String()].randomID)
	//fmt.Println(table.rows[randomID.String()].hash)
	//fmt.Println(table.rows[randomID.String()].path)	
}

// Get and remove a row with the given id
// Returns nil if the block wasn't found
// Untested and unlikely to work as intended
func (table *StoreTable) Pop(randomId *KademliaID) *storepath {

	table.lock.Lock()
	defer table.lock.Unlock()
	item := table.rows[randomId.String()]
	fmt.Println("SUCCESS for fuck sake!!")
	fmt.Println(item)
	delete(table.rows, randomId.String())
	return item
}