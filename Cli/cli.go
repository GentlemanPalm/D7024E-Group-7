package main


import (
    //"fmt"
    "flag"
)

func main() {

    textPtr := flag.String("text", "", "Text to parse.")
    ipPtr := flag.String("ip", "", "IP adress for destination.")
    flag.Parse()

    //fmt.Printf("textPtr: %s, metricPtr: %s", *textPtr, *ipPtr)

}


