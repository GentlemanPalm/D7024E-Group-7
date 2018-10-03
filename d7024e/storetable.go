package d7024e

// This file intends to implement a table of ping calls and their timeouts

import (
	"sync"
	"fmt"
	"io/ioutil"
	"log"
	//"d7024e/kademliaid"
)


type storepath struct { 
	republish bool
	pin bool
}

type StoreTable struct {
	rows map[string]*storepath
	lock *sync.Mutex
}

func NewStoreTable() *StoreTable {
	table := &StoreTable{}
	table.rows = make(map[string]*storepath)
	table.lock = &sync.Mutex{}
	return table
}

func (table *StoreTable) Push(content []byte , hash string, republish bool , pin bool) bool{
	rpc := &storepath{republish, pin}
	table.lock.Lock()
	defer table.lock.Unlock()
	table.rows[hash] = rpc

	filePath := "Files/" + hash
  err := ioutil.WriteFile(filePath, content, 0644)
  if err != nil {
		log.Fatal(err)
	}else {
		fmt.Println("------VALUES AT INDEX : --------------")
		fmt.Println(table.rows[hash].republish)
		fmt.Println(table.rows[hash].pin)
		fmt.Println("-------------------------------------")
		return true
	}	
	return false
}

func (table *StoreTable) Get(hash string) []byte {

	table.lock.Lock()
	defer table.lock.Unlock()
	item := table.rows[hash]
	if item == nil{
		return nil
	}
	content, err := ioutil.ReadFile("Files/" + hash)
	if err != nil {
		log.Fatal(err)
	}

	return content
}

func (table *StoreTable) GetNodesForRepublishing() map[string][]byte {

	table.lock.Lock()
	defer table.lock.Unlock()

	m := make(map[string][]byte)

	for k,v := range table.rows {
		if v.pin == true {
			content, err := ioutil.ReadFile("Files/" + k)
			if err != nil {
				log.Fatal(err)
				return nil
			}
			m[k] = content
		}	 
	}
	return m
}

func (table *StoreTable) Delete(hash string) []byte {

	table.lock.Lock()
	defer table.lock.Unlock()
	item := table.rows[hash]
	if item == nil{
		return nil
	}
	content, err := ioutil.ReadFile("Files/" + hash)
	if err != nil {
		log.Fatal(err)
	}

	delete(table.rows, hash)
	return content
}
