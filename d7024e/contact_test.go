package d7024e

import (
	"fmt"
	"testing"
)

func TestContacts(t *testing.T) {
	zeros := "0000000000000000000000000000000000000000"
	as := "AA000000000000000000000000000000000000AA"

	zID := NewKademliaID(zeros)
	aID := NewKademliaID(as)

	cz := NewContact(zID, "0.0.0.1")
	ca := NewContact(aID, "0.0.0.2")

	ca.CalcDistance(cz.ID)
	cz.CalcDistance(cz.ID)

	if !ca.distance.Equals(ca.ID) {
		t.Error("Distance between a contact and '0' should be the value of the contact.")
	}
	if !cz.Less(&ca) {
		t.Error("Zero vector should be less than any other vector")
	}
	//fmt.Println("Zero contact " + cz.String())
}

func TestCandidates(t *testing.T) {
	candidates := &ContactCandidates{}
	candidates.contacts = make([]Contact, 0)

	zeros := "0000000000000000000000000000000000000000"
	as := "AA000000000000000000000000000000000000AA"

	zID := NewKademliaID(zeros)
	aID := NewKademliaID(as)

	cz := NewContact(zID, "0.0.0.1")
	ca := NewContact(aID, "0.0.0.2")

	ca.CalcDistance(cz.ID)
	cz.CalcDistance(cz.ID)

	candidates.Append([]Contact{ca, cz})

	if candidates.Len() != 2 {
		t.Error("Should have 2 allocated elements")
	}

	fmt.Println(candidates.contacts)

	candidates.Sort()     // cz, ca
	candidates.Swap(0, 1) // ca, cz

	if candidates.GetContacts(2)[0].String() != ca.String() {
		t.Error("ca should be the 0th element")
	}

	if candidates.Less(0, 1) {
		t.Error("ca should not be less than cz")
	}
}
