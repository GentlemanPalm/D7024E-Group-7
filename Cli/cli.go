package main


import (
    "bufio"
    "fmt"
    "os"
    "strings"
    p "D7024E-Group-7/d7024e"
    //"flag"
)

func main() {

	var cmd, arg string
  for {
    buf := bufio.NewReader(os.Stdin)
    fmt.Print("> $ ")
    sentence, err := buf.ReadBytes('\n')
    if err != nil {
      	fmt.Println(err)
   	} else {
    	words := strings.Fields(string(sentence))
    	if(len(words) >= 2){
    		cmd = words[0]
    		arg = words[1]
    		callRPC(cmd,arg)
    	} else{
    		fmt.Println("Commands: [store , pin , unpin , pin , cat , -help]" + "\n" + "Flags: []")
    	}
    		 
    }
  }
}

func callRPC(cmd string , arg string){
	switch cmd {
    case "store":
      fmt.Println(cmd + " is about to happend, with arg: " + arg)
      hash := p.Hash(arg)
      fmt.Println(hash)
    case "pin":
      fmt.Println(cmd + " is about to happend, with arg: " + arg)
    case "unpin":
      fmt.Println(cmd + " is about to happend, with arg: " + arg)
    case "cat":
      fmt.Println(cmd + " is about to happend, with arg: " + arg)
    default:
      fmt.Println("Syntax error" + "\n" + "commands: store , pin , unpin , pin , cat ")  
  }
}
 


