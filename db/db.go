package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"syscall"
)

type ReturnChanMessage struct {
	Err  error
	Json string
}

type ReadChanMessage struct {
	Key        string
	ReturnChan chan ReturnChanMessage
}

type WriteChanMessage struct {
	Key        string
	Value      string
	ReturnChan chan ReturnChanMessage
}

type DB struct {
	data      []byte
	dataMap   map[string]string
	Log       *Log
	fd        int
	filename  string
	file      *os.File
	WriteChan chan WriteChanMessage
	ReadChan  chan ReadChanMessage
}

func (db *DB) load() {
	err := json.Unmarshal(db.data, &db.dataMap)
	if err != nil {
		fmt.Println("Error unmarshalling initial data into map: ", err)
	}
	fmt.Println(db.dataMap)
}

func (db *DB) mmap(size int) {
	fmt.Println("mmapping db file: ", size)
	data, err := syscall.Mmap(db.fd, 0, size, syscall.PROT_WRITE|syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		fmt.Println("Error mmapping: ", err)
	}
	db.data = data
}

func (db *DB) resize(size int) {
	fmt.Println("Resizing db file: ", size)
	err := syscall.Ftruncate(db.fd, int64(size))
	if err != nil {
		fmt.Println("Error resizing: ", err)
	}
}

func (db *DB) open() {
	fmt.Println("Getting db file descriptor")
	f, err := os.OpenFile(db.filename, os.O_CREATE|os.O_RDWR, 0)
	if err != nil {
		fmt.Println("Could not open file: ", err)
	}
	db.fd = int(f.Fd())
	db.file = f
}

func (db *DB) extend(size int) {
	db.file.Close()
	db.open()
	db.resize(size)
	db.mmap(size)
}

func (db *DB) write(key string, value string, returnChan chan ReturnChanMessage) {
	db.dataMap[key] = value
	b, err := json.Marshal(db.dataMap)
	if err != nil {
		fmt.Println("Error marshalling db: ", err)
	}
	if len(b) > len(db.data) {
		db.extend(len(b))
	}
	copy(db.data, b)
	m := &ReturnChanMessage{Err: nil, Json: value}
	returnChan <- *m
}

func (db *DB) listener() {
	for {
		select {
		case writeReq := <-db.WriteChan:
			key := writeReq.Key
			value := writeReq.Value
			returnChan := writeReq.ReturnChan
			db.write(key, value, returnChan)

		case readReq := <-db.ReadChan:
			key := readReq.Key
			returnChan := readReq.ReturnChan
			if value, ok := db.dataMap[key]; ok {
				m := &ReturnChanMessage{nil, value}
				returnChan <- *m
			} else {
				m := &ReturnChanMessage{errors.New("Invalid Key"), ""}
				returnChan <- *m
			}
		}
	}
}

func Init(dbFilename string, logFilename string) *DB {
	log := initLog(logFilename)
	writeChan := make(chan WriteChanMessage)
	readChan := make(chan ReadChanMessage)
	dataMap := make(map[string]string)
	var data []byte
	var fd int
	var file *os.File
	db := &DB{data, dataMap, log, fd, dbFilename, file, writeChan, readChan}
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
