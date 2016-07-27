package db

import (
	"encoding/json"
	"fmt"
	"os"
	"syscall"
)

type Entry struct {
	Command string
	Key     string
	Value   string
	Term    int
}

type Log struct {
	data     []byte
	Entries  []Entry
	fd       int
	filename string
	file     *os.File
}

func (log *Log) load() {
	err := json.Unmarshal(log.data, &log.Entries)
	if err != nil {
		fmt.Println("Error unmarshalling initial data into map: ", err)
	}
}

func (log *Log) mmap(size int) {
	fmt.Println("mmapping log file: ", size)
	data, err := syscall.Mmap(log.fd, 0, size, syscall.PROT_WRITE|syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		fmt.Println("Error mmapping: ", err)
	}
	log.data = data
}

func (log *Log) resize(size int) {
	fmt.Println("Resizing log file: ", size)
	err := syscall.Ftruncate(log.fd, int64(size))
	if err != nil {
		fmt.Println("Error resizing log file: ", err)
	}
}

func (log *Log) open() {
	fmt.Println("Getting log file descriptor")
	f, err := os.OpenFile(log.filename, os.O_CREATE|os.O_RDWR, 0)
	if err != nil {
		fmt.Println("Could not open log file: ", err)
	}
	log.fd = int(f.Fd())
	log.file = f
}

func (log *Log) extend(size int) {
	log.file.Close()
	log.open()
	log.resize(size)
	log.mmap(size)
}

func (log *Log) SetEntries(entries []Entry) {
	log.Entries = entries
	b, err := json.Marshal(log.Entries)
	if err != nil {
		fmt.Println("Error marshalling log: ", err)
	}
	if len(b) > len(log.data) {
		log.extend(len(b))
	}
	copy(log.data, b)
}

func (log *Log) AppendEntry(entry Entry) {
	log.Entries = append(log.Entries, entry)
	b, err := json.Marshal(log.Entries)
	if err != nil {
		fmt.Println("Error marshalling log: ", err)
	}
	if len(b) > len(log.data) {
		log.extend(len(b))
	}
	copy(log.data, b)
}

func initLog(filename string) *Log {
	var data []byte
	var entries []Entry
	var fd int
	var file *os.File
	log := &Log{data, entries, fd, filename, file}
	log.open()
	f, err := os.Stat(filename)
	if err != nil {
		fmt.Println("Could not stat file: ", err)
	}
	size := int(f.Size())
	if size == 0 {
		size = 10
		log.resize(size)
	}
	log.mmap(size)
	return log
}
