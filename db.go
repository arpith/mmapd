package main

import (
	"encoding/json"
	"fmt"
	"os"
	"syscall"
)

type readChanMessage struct {
	key        string
	returnChan chan string
}

type db struct {
	data      []byte
	dataMap   map[string]string
	log       *log
	fd        int
	filename  string
	file      *os.File
	writeChan chan map[string]string
	readChan  chan readChanMessage
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

func (db *db) listener() {
	for {
		select {
		case writeReq := <-db.writeChan:
			key := writeReq["key"]
			value := writeReq["value"]
			db.dataMap[key] = value
			b, err := json.Marshal(db.dataMap)
			if err != nil {
				fmt.Println("Error marshalling db: ", err)
			}
			if len(b) > len(db.data) {
				db.extend(len(b))
			}
			copy(db.data, b)
			s := "SET " + key + ": " + value
			b = append(db.log.data, []byte(s)...)
			if len(b) > len(db.log.data) {
				db.log.extend(len(b))
			}
			copy(db.log.data, b)

		case readReq := <-db.readChan:
			key := readReq.key
			readReq.returnChan <- db.dataMap[key]
		}
	}
}

func initDB(dbFilename string, logFilename string) *db {
	log := initLog(logFilename)
	writeChan := make(chan map[string]string, 250)
	readChan := make(chan readChanMessage)
	dataMap := make(map[string]string)
	var data []byte
	var fd int
	var file *os.File
	db := &db{data, dataMap, log, fd, dbFilename, file, writeChan, readChan}
	db.open()
	f, err := os.Stat(dbFilename)
	if err != nil {
		fmt.Println("Could not stat file: ", err)
	}

	db.mmap(int(f.Size()))
	db.load()
	go db.listener()
	return db
}
