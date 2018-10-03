package d7024e

// This file intends to implement a table of ping calls and their timeouts

import (
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"
	//"d7024e/kademliaid"
)

type storepath struct {
	republish bool
	pin       bool
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

func (table *StoreTable) Push(content []byte, hash string, republish bool, pin bool) bool {
	rpc := &storepath{republish, pin}
	table.lock.Lock()
	defer table.lock.Unlock()

	if table.rows[hash] != nil {
		table.rows[hash].republish = true
		return true
	} else {
		table.rows[hash] = rpc

		filePath := "Files/" + hash
		err := ioutil.WriteFile(filePath, content, 0644)
		if err != nil {
			log.Fatal(err)
		} else {
			fmt.Println("Received file with hash " + hash)
			return true
		}
		return false
	}
}

func (table *StoreTable) Get(hash string) []byte {

	table.lock.Lock()
	defer table.lock.Unlock()
	item := table.rows[hash]
	if item == nil {
		return nil
	}
	content, err := ioutil.ReadFile("Files/" + hash)
	if err != nil {
		log.Fatal(err)
	}

	return content
}

func (st *StoreTable) Expire() {
	time.Sleep(time.Duration(30) * time.Second)

	st.lock.Lock()
	defer st.lock.Unlock()

	toDelete := make(map[string]bool)

	for k, v := range st.rows {
		if v.republish || v.pin {
			v.republish = false
		} else {
			fmt.Println("Deleted " + k + " because of lack of republishing")
			toDelete[k] = true
		}
	}

	for k, _ := range toDelete {
		delete(st.rows, k)
		fmt.Println("Deltd fo real")
		// TODO: Delete file?
	}

	go st.Expire()
}

func (table *StoreTable) GetNodesForRepublishing() map[string][]byte {

	table.lock.Lock()
	defer table.lock.Unlock()

	m := make(map[string][]byte)

	for k, v := range table.rows {
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
	if item == nil {
		return nil
	}
	content, err := ioutil.ReadFile("Files/" + hash)
	if err != nil {
		log.Fatal(err)
	}

	delete(table.rows, hash)
	return content
}
