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
    		SendRequest(cmd,arg)
    	} else{
    		fmt.Println("Commands: [store , pin , unpin , pin , cat , -help]" + "\n" + "Flags: []")
    	}
    		 
    }
  }
}

func SendRequest(rpc string , arg string){
  
  client := &http.Client{}
  reqStruct := &Request{}

  url := "http://localhost:8080" + "/" + rpc + "/"

  switch rpc {
    case "store":
      fmt.Println(rpc + " is about to happend, with arg: " + arg)
      value := readFile(arg)
      content := base64.StdEncoding.EncodeToString(value)

      reqStruct.Hash = hash(arg)
      reqStruct.Content = content

    case "pin":
      fmt.Println(rpc + " is about to happend, with arg: " + arg)
      reqStruct.Hash = arg
      reqStruct.Content = ""

    case "unpin":
      fmt.Println(rpc + " is about to happend, with arg: " + arg)
      reqStruct.Hash = arg
      reqStruct.Content = ""

    case "cat":
      fmt.Println(rpc + " is about to happend, with arg: " + arg)
      reqStruct.Hash = arg
      reqStruct.Content = ""

    default:
      fmt.Println("Syntax error" + "\n" + "commands: store , pin , unpin , pin , cat ")  
  }

  c := strings.NewReader(marshalRequest(reqStruct))

  req, err := http.NewRequest("POST", url, c)
  if err != nil {
    log.Fatalln(err)
  }else{
    req.Header.Set("Content-Type", "application/json")

    resp, err1 := client.Do(req)
    if err1 != nil {
      log.Fatalln(err1)
    }

    response := parseRequest(resp)
    responsePrint(response, rpc)

  }
}

func responsePrint(response *Response , rpc string){
  switch rpc {
    case "store":
      fmt.Println(rpc + " request status: " + response.Status)
      fmt.Println(rpc + " Hash: " + response.Content)
    case "pin":
      fmt.Println(rpc + " request status: " + response.Status)
    case "unpin":
      fmt.Println(rpc + " request status: " + response.Status)
    case "cat":
      con := readFile(response.Content)
      fmt.Println(rpc + " request status: " + response.Status)
      fmt.Println("Content: ")
      fmt.Println(con)
    default:
      fmt.Println("Error rpc did not succed")  
  }
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

func parseRequest(r *http.Response) *Response {
  body, err := ioutil.ReadAll(r.Body)
  if err != nil {
    fmt.Println(err)
  }
  data := &Response{}
  result := json.Unmarshal(body, data)
  if result != nil {
    fmt.Println(result)
  }
  return data
}
 


