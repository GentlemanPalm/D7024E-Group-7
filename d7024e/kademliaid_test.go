package d7024e

import "testing"

func TestKademliaID(t *testing.T) {
	rand := NewRandomKademliaID()

	if rand.Equals(NewRandomKademliaID()) {
		t.Error("Two random kademlia IDs should not be the same. The chance of this happening is less than one in a billion.")
	}

	if !rand.Equals(rand) {
		t.Error("A Kademlia ID should be equal to itself.")
	}

	zeros := "0000000000000000000000000000000000000000"
	as := "AA000000000000000000000000000000000000AA"

	zID := NewKademliaID(zeros)
	aID := NewKademliaID(as)

	if zID.String() != zeros {
		t.Error("String representation of Kademlia ID is mangled.")
	}

	if !zID.CalcDistance(aID).Equals(aID) {
		t.Error("XOR property does not seem to hold")
	}

}
