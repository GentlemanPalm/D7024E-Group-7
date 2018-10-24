package d7024e

import (
	"testing"
	"fmt"

)

func TestHashFile(t *testing.T) {
	filePath := "text.txt"
	expected := "0d3fd7c45c3a00e740f584098e825c56ab731bf5"
	hash := Hash(filePath)

	fmt.Println(expected)
	fmt.Println(hash)

	if expected != hash{
		t.Error("Hash do not match")
	}
}