// This file contains a table which is meant to keep track of which
// node lookup request a certain node lookup resonse is referring.

// This is to facilitate the iterative NodeLookup procedure
package d7024e

import (
	"sync"
)

type NodeLookupTable struct {
	rows map[string]*Shortlist
	lock *sync.Mutex
}

func NewNodeLookupTable() *NodeLookupTable {
	table := &NodeLookupTable{}
	table.rows = make(map[string]*Shortlist)
	table.lock = &sync.Mutex{}
	return table
}

func (table *NodeLookupTable) Put(randomID *KademliaID, shortlist *Shortlist) {
	table.lock.Lock()
	defer table.lock.Unlock()
	table.rows[randomID.String()] = shortlist
}

// Get and remove an item from the table, return nil if item not found
func (table *NodeLookupTable) Pop(randomID *KademliaID) *Shortlist {
	table.lock.Lock()
	defer table.lock.Unlock()
	item := table.rows[randomID.String()]
	if item != nil {
		delete(table.rows, randomID.String())
	}
	return item
}
