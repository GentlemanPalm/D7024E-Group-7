package d7024e

import (
	"container/list"
	"sync"
	"fmt"
)

// TODO: Make buckets no larger than k in size

// bucket definition
// contains a List
type bucket struct {
	list *list.List
	lock *sync.Mutex
}

// newBucket returns a new instance of a bucket
func newBucket() *bucket {
	bucket := &bucket{}
	bucket.list = list.New()
	bucket.lock = &sync.Mutex{}
	return bucket
}

func (bucket *bucket) AddContact(contact Contact, network *Network) {
	bucket.lock.Lock()
	defer bucket.lock.Unlock()
	bucket.addContact(contact,network)
}

func (bucket *bucket) ReplaceContact(old *KademliaID, replacement *Contact , network *Network) {
	bucket.lock.Lock()
	defer bucket.lock.Unlock()

	for e := bucket.list.Front(); e != nil; e = e.Next() {
		contact := e.Value.(Contact)

		if (contact).ID.Equals(old) {
			bucket.list.Remove(e)
			break
		}
	}

	bucket.addContact(*replacement, network)
}

// AddContact adds the Contact to the front of the bucket
// or moves it to the front of the bucket if it already existed
func (bucket *bucket) addContact(contact Contact , network *Network) {
	var element *list.Element
	if bucket.list.Len() > bucketSize {
		fmt.Println("-------------Bucket Full-------------")
		fmt.Println(bucket.list.Len())
		c := bucket.list.Back().Value.(Contact)
		oldContact := &c
		newContact := &contact
		network.SendPingMessageWithReplacement(oldContact, newContact)
	}else{
		fmt.Println("-------------Bucket size-------------")
		fmt.Println(bucket.list.Len())
		for e := bucket.list.Front(); e != nil; e = e.Next() {
			nodeID := e.Value.(Contact).ID

			if (contact).ID.Equals(nodeID) {
				element = e
			}
		}

		if element == nil {
			if bucket.list.Len() < bucketSize {
				bucket.list.PushFront(contact)
			}
		} else {
			bucket.list.MoveToFront(element)
		}
	}
	
}

// GetContactAndCalcDistance returns an array of Contacts where
// the distance has already been calculated
func (bucket *bucket) GetContactAndCalcDistance(target *KademliaID) []Contact {
	bucket.lock.Lock()
	defer bucket.lock.Unlock()
	return bucket.getContactAndCalcDistance(target)
}

// GetContactAndCalcDistance returns an array of Contacts where
// the distance has already been calculated
func (bucket *bucket) getContactAndCalcDistance(target *KademliaID) []Contact {
	var contacts []Contact

	for elt := bucket.list.Front(); elt != nil; elt = elt.Next() {
		contact := elt.Value.(Contact)
		contact.CalcDistance(target)
		contacts = append(contacts, contact)
	}

	return contacts
}

// Len return the size of the bucket
func (bucket *bucket) Len() int {
	return bucket.list.Len()
}
