package main

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"os"
	"syscall"
)

type db struct {
	data      []byte
	dataMap   map[string]string
	fd        int
	filename  string
	file      *os.File
	writeChan chan map[string]string
}

func (db *db) load() {
	err := json.Unmarshal(db.data, &db.dataMap)
	if err != nil {
		fmt.Println("Error unmarshalling initial data into map: ", err)
	}
	fmt.Println(db.dataMap)
}

func (db *db) mmap(size int) {
	fmt.Println("mmapping: ", size)
	data, err := syscall.Mmap(db.fd, 0, size, syscall.PROT_WRITE|syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		fmt.Println("Error mmapping: ", err)
	}
	db.data = data
}

func (db *db) resize(size int) {
	fmt.Println("Resizing: ", size)
	err := syscall.Ftruncate(db.fd, int64(size))
	if err != nil {
		fmt.Println("Error resizing: ", err)
	}
}

func (db *db) open() {
	fmt.Println("Getting file descriptor")
	f, err := os.OpenFile(db.filename, os.O_CREATE|os.O_RDWR, 0)
	if err != nil {
		fmt.Println("Could not open file: ", err)
	}
	db.fd = int(f.Fd())
	db.file = f
}

func (db *db) extend(size int) {
	db.file.Close()
	db.open()
	db.resize(size)
	db.mmap(size)
}

func (db *db) writer() {
	for {
		req := <-db.writeChan
		db.dataMap[req["key"]] = req["value"]
		b, err := json.Marshal(db.dataMap)
		if err != nil {
			fmt.Println("Error marshalling db: ", err)
		}
		if len(b) > len(db.data) {
			db.extend(len(b))
		}
		copy(db.data, b)
	}
}

func (db *db) handler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	switch r.Method {
	case "GET":
		key := ps.ByName("key")
		value := db.dataMap[key]
		fmt.Fprintln(w, value)
	case "POST":
		m := make(map[string]string)
		m["key"] = ps.ByName("key")
		m["value"] = r.FormValue("value")
		db.writeChan <- m
		fmt.Fprintln(w, r.FormValue("value"))
	}
}

func NewHandler(db *db) func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	return db.handler
}

func openDB(filename string) *db {
	writeChan := make(chan map[string]string, 250)
	dataMap := make(map[string]string)
	var data []byte
	var fd int
	var file *os.File
	db := &db{data, dataMap, fd, filename, file, writeChan}
	db.open()
	f, err := os.Stat(filename)
	if err != nil {
		fmt.Println("Could not stat file: ", err)
	}

	db.mmap(int(f.Size()))
	db.load()
	go db.writer()
	return db
}
