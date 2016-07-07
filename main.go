package main

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"os"
	"strings"
	"syscall"
)

type db struct {
	data      []byte
	dataMap   map[string]string
	fd        int
	writeChan chan map[string]string
}

func (db *db) remap(size int) {
	fmt.Println("Remapping: ", size*2)
	data, err := syscall.Mmap(db.fd, 0, size*2, syscall.PROT_WRITE|syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		fmt.Println("Error remapping: ", err)
	}
	db.data = data
}

func (db *db) resize(size int) {
	fmt.Println("Resizing: ", size*2)
	err := syscall.Ftruncate(db.fd, int64(size*2))
	if err != nil {
		fmt.Println("Error resizing: ", err)
	}
}

func (db *db) writer() {
	for {
		req := <-db.writeChan
		fmt.Println("DB before modification: ", string(db.data))
		db.dataMap[req["key"]] = req["value"]
		b, err := json.Marshal(db.dataMap)
		if err != nil {
			fmt.Println("Error marshalling db: ", err)
		}
		if len(b) > len(db.data) {
			fmt.Println("Need to resize!")
			db.resize(len(b))
			db.remap(len(b))
		}
		copy(db.data, b)
		fmt.Println("DB after modification: ", string(db.data))
	}
}

func (db *db) get(key string) string {
	return db.dataMap[key]
}

func (db *db) handler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	switch r.Method {
	case "GET":
		key := ps.ByName("key")
		fmt.Printf("Getting %s!\n", key)
		fmt.Fprintln(w, db.get(key))
	case "POST":
		fmt.Println("Got request")
		m := make(map[string]string)
		m["key"] = ps.ByName("key")
		m["value"] = r.FormValue("value")
		fmt.Println(m)
		db.writeChan <- m
		fmt.Fprintln(w, r.FormValue("value"))
	}
}

func NewHandler(db *db) func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	return db.handler
}

func openDB() *db {
	filename := "db"

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0)
	if err != nil {
		fmt.Println("Could not open file: ", err)
	}

	fi, err := os.Stat(filename)
	if err != nil {
		fmt.Println("Could not stat file: ", err)
	}

	data, err := syscall.Mmap(int(f.Fd()), 0, int(fi.Size()), syscall.PROT_WRITE|syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		fmt.Printf("Could not memory map the file (into a byte array): ", err)
	}

	var dataMap map[string]string

	err = json.Unmarshal(data, &dataMap)
	if err != nil {
		fmt.Println("Error unmarshalling initial data into map: ", err)
	}

	writeChan := make(chan map[string]string)
	m := &db{data, dataMap, int(f.Fd()), writeChan}
	go m.writer()
	return m
}

func main() {

	m := openDB()
	handler := NewHandler(m)

	router := httprouter.New()
	router.GET("/get/:key", handler)
	router.POST("/set/:key", handler)

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "3001"
	}

	http.ListenAndServe(":"+port, router)
}
