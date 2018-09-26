package d7024e

// This is the implementation of the 'shortlist' used by the Node Lookup procedure.

import (
	"sort"
	"sync"
	//"d7024e/kademliaid"
)

type ShortlistItem struct {
	contact *Contact
	visited uint // 0 for not yet visited, 1 for request sent, 2 for response received
	lock    *sync.Mutex
}

/*
 * The actual shortlist, thread safe.
 * */
type Shortlist struct {
	items  []ShortlistItem // Items currently in consideration
	dead   []ShortlistItem // Items verified to be dead
	target *KademliaID
	me     *Contact
	lock   *sync.Mutex
}

/*
 * The table concerning multiple Shortlists for node lookups.
 * Might not be necessary for the program, as it might be possible to just keep
 * passing references of shortlists
 * */
type ShortlistTable struct {
	lists []Shortlist
	lock  *sync.Mutex
}

func NewShortlistTable() *ShortlistTable {
	table := &ShortlistTable{}
	table.lists = make([]Shortlist, 20) // TODO: Make global variable for K
	table.lock = &sync.Mutex{}
	return table
}

// Can never _ever_ add oneself as a contact
func NewShortlist(me *Contact, target *KademliaID) *Shortlist {
	shortlist := &Shortlist{}
	shortlist.target = target
	shortlist.items = make([]ShortlistItem, 20)
	shortlist.lock = &sync.Mutex{}
	return shortlist
}

func NewShortlistItem(contact *Contact) *ShortlistItem {
	si := &ShortlistItem{}
	si.contact = contact
	si.visited = 0
	si.lock = &sync.Mutex{}
	return si
}

/*
 * Gets the closest unvisited node and marks it as visited or nil.
 * The nil value implies there are no unvisited nodes in the
 * current list, thus signaling that it should conclude.
 * */
func (shortlist *Shortlist) GetClosestUnvisited() *Contact {
	shortlist.lock.Lock()
	defer shortlist.lock.Unlock()
	for i := 0; i < len(shortlist.items); i++ {
		if shortlist.items[i].visited == 0 && shortlist.items[i].contact != nil {
			shortlist.items[i].visited = 1
			return shortlist.items[i].contact
		}
	}
	return nil
}

/*
 * Takes a contact and moves it to the dead list. Items on the
 * 'dead' list won't be considered in the future.
 * */
func (shortlist *Shortlist) MarkAsDead(contact *Contact) {
	shortlist.lock.Lock()
	defer shortlist.lock.Unlock()

	for i := 0; i < len(shortlist.items); i++ {
		// Item isn't empty
		if !shortlist.items[i].isAvailable() {
			// Visited but not responded
			if shortlist.items[i].contact.ID.Equals(contact.ID) && shortlist.items[i].visited == 1 {
				// Add to dead and remove item from list
				shortlist.dead = append(shortlist.dead, shortlist.items[i])
				shortlist.items[i] = ShortlistItem{}
				shortlist.sort()
				return
			}
		}
	}
	return
}

/*
 * Adds all contacts in the list to the list of considered, as long as they aren't
 * */
func (shortlist *Shortlist) AddContacts(contacts []Contact) {
	shortlist.lock.Lock()
	defer shortlist.lock.Unlock()
	shortlist.addContacts(contacts)
}

func (shortlist *Shortlist) addContacts(contacts []Contact) {
	for i := 0; i < len(contacts); i++ {
		if shortlist.notInList(&contacts[i]) {
			shortlist.addContactIfSufficientlyClose(&contacts[i])
			shortlist.sort()
		}
	}
	shortlist.sort()
}

func (shortlist *Shortlist) addContactIfSufficientlyClose(contact *Contact) {
	if contact == nil || contact.ID == nil {
		return
	}
	distance := contact.ID.CalcDistance(shortlist.target)

	index := shortlist.getFirstAvailableIndexOrEnd()
	item := shortlist.items[index]

	// Check is a) item is empty or b) item is further away than the new item
	if (&item).isAvailable() {
		shortlist.items[index] = *NewShortlistItem(contact)
	} else if distance.Less(item.contact.ID.CalcDistance(shortlist.target)) {
		shortlist.items[index] = *NewShortlistItem(contact)
	}
}

func (itm *ShortlistItem) isAvailable() bool {
	return itm == nil || itm.contact == nil || itm.contact.ID == nil
}

func (shortlist *Shortlist) getFirstAvailableIndexOrEnd() int {
	for i := 0; i < len(shortlist.items); i++ {
		if shortlist.items[i].contact == nil || shortlist.items[i].contact.ID == nil {
			return i
		}
	}
	return len(shortlist.items) - 1
}

// Helper function to determine that the
func (shortlist *Shortlist) notInList(contact *Contact) bool {
	if contact.ID.Equals(shortlist.me.ID) {
		return false // The 'me' ID is never valid
	}
	for i := 0; i < len(shortlist.dead); i++ {
		if shortlist.dead[i].contact != nil {
			if shortlist.dead[i].contact.ID.Equals(contact.ID) {
				return false
			}
		}
	}
	for i := 0; i < len(shortlist.items); i++ {
		if shortlist.items[i].contact != nil {
			if shortlist.items[i].contact.ID.Equals(contact.ID) {
				return false
			}
		}
	}
	return true
}

// Sort the Contacts in ContactCandidates
func (shortlist *Shortlist) sort() {
	//sort.Sort(shortlist)
	sort.Slice(shortlist.items, func(i, j int) bool {
		if shortlist.items[i].contact == nil && shortlist.items[j].contact == nil {
			return false
		}
		if shortlist.items[i].contact == nil {
			return true
		}
		if shortlist.items[j].contact == nil {
			return false
		}

		t1 := shortlist.items[i].contact.ID.CalcDistance(shortlist.target)
		t2 := shortlist.items[j].contact.ID.CalcDistance(shortlist.target)

		return t1.Less(t2)
	})
}

/*
// Len returns the length of the ContactCandidates
func (shortlist *Shortlist) Len() int {
	return len(shortlist.items)
}

// Swap the position of the items at i and j
func (shortlist *Shortlist) Swap(i, j int) {
	shortlist.items[i], shortlist.items[j] = shortlist.items[j], shortlist.items[i]
}

// Less returns true if the Contact at index i is smaller than
// the Contact at index j
func (shortlist *Shortlist) Less(i, j int) bool {
	return shortlist.items[i].Less(&shortlist.items[j])
}

func (si *ShortlistItem) Less(sj *ShortlistItem) bool {
	return si.contact.Less(sj.contact)
}*/
