package d7024e

import (
	"fmt"
	"crypto/sha1"
	"os"
	"io/ioutil"
	"log"
)

func Hash(arg string) {

	content, err := ioutil.ReadFile(arg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("File contents: %s", content)
	
	file, err := os.Open(arg) 
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", "\n")
	fmt.Printf("File: %s", file)

	 h := sha1.New()
	 h.Write([]byte(content))
	 bs := h.Sum(nil)

	 fmt.Printf("%s", "\n" + "Content: ")
	 fmt.Printf("%s", content)
	 fmt.Printf("%s", "\n" + "Hash: ")
	 fmt.Printf("%x\n",bs)

	 file.Close()
}
