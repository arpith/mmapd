package main

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"os"
	"syscall"
)

type log struct {
	data      []byte
	fd        int
	filename  string
	file      *os.File
	writeChan chan map[string]string
}

func (log *log) mmap(size int) {
	fmt.Println("mmapping log file: ", size)
	data, err := syscall.Mmap(log.fd, 0, size*2, syscall.PROT_WRITE|syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		fmt.Println("Error mmapping: ", err)
	}
	log.data = data
}

func (log *log) resize(size int) {
	fmt.Println("Resizing log file: ", size)
	err := syscall.Ftruncate(log.fd, int64(size*2))
	if err != nil {
		fmt.Println("Error resizing log file: ", err)
	}
}

func (log *log) open() {
	fmt.Println("Getting log file descriptor")
	f, err := os.OpenFile(log.filename, os.O_CREATE|os.O_RDWR, 0)
	if err != nil {
		fmt.Println("Could not open log file: ", err)
	}
	log.fd = int(f.Fd())
	log.file = f
}

func (log *log) extend(size int) {
	log.file.Close()
	log.open()
	log.resize(size)
	log.mmap(size)
}

func initLog(filename string) *log {
	var data []byte
	var fd int
	var file *os.File
	log := &log{data, fd, filename, file}
	log.open()
	f, err := os.Stat(filename)
	if err != nil {
		fmt.Println("Could not stat file: ", err)
	}
	log.mmap(int(f.Size()))
	log.load()
	return log
}
