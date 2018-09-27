package d7024e

// This is the implementation of the 'shortlist' used by the Node Lookup procedure.

import (
	"NetworkMessage"
	"fmt"
	"strconv"
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
	items    map[string]*ShortlistItem // Items currently in consideration
	dead     map[string]*ShortlistItem // Items verified to be dead
	target   *KademliaID
	me       *Contact
	callback NodeLookupCallback
	lock     *sync.Mutex
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
func NewShortlist(me *Contact, target *KademliaID, callback NodeLookupCallback) *Shortlist {
	shortlist := &Shortlist{}
	shortlist.target = target
	shortlist.me = me
	shortlist.items = make(map[string]*ShortlistItem) //make([]ShortlistItem, 20)
	shortlist.dead = make(map[string]*ShortlistItem)
	shortlist.callback = callback
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

func CloneMap(oldMap map[string]*ShortlistItem) map[string]*ShortlistItem {
	newMap := make(map[string]*ShortlistItem)
	for k, v := range oldMap {
		newMap[k] = v
	}
	return newMap
}

func (shortlist *Shortlist) GetClosestUnvisited() *Contact {
	shortlist.lock.Lock()
	defer shortlist.lock.Unlock()
	return shortlist.getClosestUnvisited()
}

/*
 * Gets the closest unvisited node and marks it as visited or nil.
 * The nil value implies there are no unvisited nodes in the
 * current list, thus signaling that it should conclude.
 * */
func (shortlist *Shortlist) getClosestUnvisited() *Contact {
	var closest *ShortlistItem
	closest = nil
	for _, value := range shortlist.items {
		if closest == nil && value.visited == 0 {
			closest = value
		} else {
			if value != nil && value.contact != nil {
				if value.visited == 0 && value.contact.ID.CalcDistance(shortlist.target).Less(closest.contact.ID.CalcDistance(shortlist.target)) {
					closest = value
				}
			}
		}
	}
	if closest == nil {
		return nil
	}
	closest.visited = 1
	return closest.contact
}

func (shortlist *Shortlist) MarkAsDead(contact *Contact) {
	shortlist.lock.Lock()
	defer shortlist.lock.Unlock()
	shortlist.markAsDead(contact)
}

/*
 * Takes a contact and moves it to the dead list. Items on the
 * 'dead' list won't be considered in the future.
 * */
func (shortlist *Shortlist) markAsDead(contact *Contact) {
	if contact.ID != nil {
		item := shortlist.items[contact.ID.String()]
		if item != nil && item.visited != 2 {
			delete(shortlist.items, contact.ID.String())
			shortlist.dead[contact.ID.String()] = item
			fmt.Println("Marked " + contact.ID.String() + " as dead.")
		} else {
			fmt.Println("Item not found or already verified as visited ")
		}
	}

}

/*
 * Adds all contacts in the list to the list of considered, as long as they aren't
 * */
func (shortlist *Shortlist) AddContacts(contacts []Contact) bool {
	shortlist.lock.Lock()
	defer shortlist.lock.Unlock()
	return shortlist.addContacts(contacts)
}

func (shortlist *Shortlist) acertainLiving(target *KademliaID) bool {
	if shortlist.isActive(target) {
		shortlist.items[target.String()].visited = 2
		return true
	} else {
		fmt.Println(target.String() + " is ded.")
	}
	return false
}

func (shortlist *Shortlist) isActive(target *KademliaID) bool {
	return shortlist.isAlive(target) && (shortlist.items[target.String()] != nil)
}

func (shortlist *Shortlist) isAlive(target *KademliaID) bool {
	_, isDed := shortlist.dead[target.String()]
	return !isDed
}

func (shortlist *Shortlist) HandleResponse(network *Network, sender *KademliaID, response *NetworkMessage.ValueResponse) {
	shortlist.lock.Lock()
	defer shortlist.lock.Unlock()
	shortlist.acertainLiving(sender) // TODO: Need to take any particular care about the sender being dead?
	fmt.Println("Entered shortlist.HandleResponse.")
	contacts := shortlist.parseResponseAsContacts(response)
	if contacts == nil {
		fmt.Println("Is this a node lookup, because this doesn't seem like a node lookup!")
	} else {
		fmt.Println("Adding contacts, there are " + strconv.Itoa(len(*contacts)) + " of them")
		shortlist.addContacts(*contacts)
	}
	shortlist.doCleanup(network)
}

func (shortlist *Shortlist) parseResponseAsContacts(message *NetworkMessage.ValueResponse) *[]Contact {
	fmt.Println("Prasing message contents for nodeLookup in hopes of accomplishing something")
	contacts := make([]Contact, 20) // TODO: Use K
	switch response := message.Response.(type) {
	case *NetworkMessage.ValueResponse_Nodes:
		nodes := response.Nodes.Nodes
		for i := range nodes { // TODO: Make it work for FIND_VALUE
			fmt.Println(nodes[i].KademliaId + " @ " + nodes[i].Address)
			kID := NewKademliaID(nodes[i].KademliaId)
			if !shortlist.me.ID.Equals(kID) {
				contacts[i] = NewContact(kID, nodes[i].Address)
			} else {
				fmt.Println("But  was sent my own ID! I can't add myself now, can I?")
			}

		}
	case *NetworkMessage.ValueResponse_Content:
		fmt.Println("Cannot handle content values just yet")
	}
	return &contacts
}

func (shortlist *Shortlist) HandleTimeout(network *Network, sender *KademliaID) {
	shortlist.lock.Lock()
	defer shortlist.lock.Unlock()
	contact := NewContact(sender, "0.0.0.0")
	shortlist.markAsDead(&contact) // TODO: Check if IP is needed
	shortlist.doCleanup(network)
}

//
func (shortlist *Shortlist) doCleanup(network *Network) {
	shortlist.launchRequests(network)
	if shortlist.hasFinished() {
		fmt.Println("No ongoing requests or unvisited data... finishing up")
		contacts := make([]Contact, 20)
		i := 0
		for _, v := range shortlist.items {
			if v != nil {
				contacts[i] = *(v.contact)
				i++
			}
		}
		if shortlist.callback != nil {
			shortlist.callback(contacts)
			shortlist.callback = nil
		}
	} else {
		fmt.Println("Hit cleanup, but still has either unvisited nodes or ongoing requests")
	}
}

func (shortlist *Shortlist) LaunchRequests(network *Network) bool {
	shortlist.lock.Lock()
	defer shortlist.lock.Unlock()
	return shortlist.launchRequests(network)
}

// Launch requests so that 3 requests are active at once
// It should have the following behavior regarding requests:
//
// If there are 3 requests already running, then do nothing.
// If there are less than three requests, launch requests equal to the difference
//     but only if there is an unvisited contact in the list
func (shortlist *Shortlist) launchRequests(network *Network) bool {
	target := 3 - shortlist.countActiveRequests() // TODO: Add global for alpha
	hasLaunched := false
	for i := 0; i < target; i++ {
		recipient := shortlist.getClosestUnvisited()
		if recipient == nil {
			fmt.Println("Tried to launch " + strconv.Itoa(target) + " requests, but there aren't enough unvisited nodes in shortlist")
			continue
		} else {
			hasLaunched = true
			fmt.Println("Sending FIND_* request to " + recipient.ID.String())
			fmt.Println("TODO: Actually send the thing")
			go network.SendFindNodeForNodeLookup(shortlist.target, recipient, shortlist)
			//go network.SendFindContactMessage() // TODO: Make message for node lokoups
		}
	}
	return hasLaunched
}

func (shortlist *Shortlist) hasUnvisited() bool {
	for _, v := range shortlist.items {
		if v != nil && v.visited == 0 {
			return true
		}
	}
	return false
}

// The NodeLookup algorithm has finished iff
// 1. There are no unvisited nodes available AND
// 2. There are no onging queries
func (shortlist *Shortlist) hasFinished() bool {
	return shortlist.countActiveRequests() == 0 && !shortlist.hasUnvisited()
}

func (shortlist *Shortlist) countActiveRequests() int {
	counter := 0
	for _, v := range shortlist.items {
		if v != nil {
			if v.visited == 1 {
				counter++
			}
		}
	}
	return counter
}

func (shortlist *Shortlist) addContacts(contacts []Contact) bool {
	oldMap := CloneMap(shortlist.items)
	for i := range contacts { // Add all the contacts
		shi := NewShortlistItem(&contacts[i])
		if shi.contact.ID != nil && shortlist.isAlive(shi.contact.ID) && shortlist.items[shi.contact.ID.String()] == nil {
			shortlist.items[shi.contact.ID.String()] = shi
		}
		//fmt.Println("----")
		//fmt.Println(shi.contact.ID.String())
		//fmt.Println(shortlist.isAlive(shi.contact.ID))
		//fmt.Println(shortlist.items[shi.contact.ID.String()])
	} // Throw away all but the k closest
	shortlist.prune()
	return !isSame(shortlist.items, oldMap)
}

func isSame(updated map[string]*ShortlistItem, outdated map[string]*ShortlistItem) bool {
	matches := 0
	for k, _ := range updated {
		if outdated[k] == nil {
			return false
		} else {
			matches++
		}
	}
	// If there are elements in the outdated list not in the updated list, they cannot be the same
	return matches == len(outdated)
}

// Removes items from the hashmap until only the k closest remains
func (shortlist *Shortlist) prune() {
	if len(shortlist.items) > 20 {
		shortlist.removeMostDistant()
		shortlist.prune()
	}
}

// Removes the most distant elements of the shortlist if required
func (shortlist *Shortlist) removeMostDistant() {
	var mostDistant *ShortlistItem
	for _, value := range shortlist.items {
		if mostDistant == nil {
			mostDistant = value
		} else {
			d1 := mostDistant.contact.ID.CalcDistance(shortlist.target)
			d2 := value.contact.ID.CalcDistance(shortlist.target)
			if d1.Less(d2) {
				mostDistant = value
			}
		}
	}
	if mostDistant != nil {
		delete(shortlist.items, mostDistant.contact.ID.String())
	}
}
