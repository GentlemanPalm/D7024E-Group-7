package main


import (
  "bufio"
  "os"
  "fmt"
  "strings"
  "encoding/json"
  "encoding/base64"
  "net/http"
  "crypto/sha1"
  "io/ioutil"
  "log"
)

type Request struct {
  Hash    string `json:"hash"`
  Content string `json"content,omitempty"`
}

// The generic response used for all requests
type Response struct {
  Status  string `json:"status"`
  Content string `json:"content,omitempty"`
}

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
      SendRequest(cmd,arg)
    case "pin":
      fmt.Println(cmd + " is about to happend, with arg: " + arg)
      SendRequest(cmd,arg)
    case "unpin":
      fmt.Println(cmd + " is about to happend, with arg: " + arg)
      SendRequest(cmd,arg)
    case "cat":
      fmt.Println(cmd + " is about to happend, with arg: " + arg)
      SendRequest(cmd,arg)  
    default:
      fmt.Println("Syntax error" + "\n" + "commands: store , pin , unpin , pin , cat ")  
  }
}

func SendRequest(rpc string , arg string){
  
  client := &http.Client{}
  value := readFile(arg)
  content := base64.StdEncoding.EncodeToString(value)

  reqStruct := &Request {
    Hash: hash(arg),
    Content: content,
  }

  c := strings.NewReader(marshalRequest(reqStruct))

  url := "http://localhost:8080" + "/" + rpc + "/"

  req, err := http.NewRequest("POST", url, c)
  if err != nil {
    log.Fatalln(err)
  }
  req.Header.Set("Content-Type", "application/json")

  resp, err1 := client.Do(req)
  if err1 != nil {
    log.Fatalln(err1)
  }

  //response := marshalResponse(resp)

  var result map[string]interface{}
  json.NewDecoder(resp.Body).Decode(&result)
  log.Println(result)
}

func marshalRequest(request *Request) string {
  marsh, merr := json.Marshal(request)

  if merr != nil {
    fmt.Println(merr)
  }

  s := string(marsh[:len(marsh)])

  return s
}

func hash(arg string) string {

  content, err := ioutil.ReadFile(arg)
  if err != nil {
    log.Fatal(err)
  }

  h := sha1.New()
  h.Write([]byte(arg))
  h.Write([]byte(content))
  bs := h.Sum(nil)

  str := fmt.Sprintf("%x\n", bs)

  return str;
}

func readFile(arg string) []byte {

  fmt.Println(arg)

  content, err := ioutil.ReadFile(arg)
  if err != nil {
    log.Fatal(err)
  }

  return content
}

 


