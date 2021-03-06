package d7024e

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"log"
)

func Hash(arg string) string {

	content, err := ioutil.ReadFile(arg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("File contents: %s", content)

	h := sha1.New()
	h.Write([]byte(arg))
	h.Write([]byte(content))
	bs := h.Sum(nil)

	fmt.Printf("%s", "\n"+"Hash: ")
	fmt.Printf("%x\n", bs)

	str := fmt.Sprintf("%x", bs)

	return str
}
