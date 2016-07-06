package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"os"
	"strings"
	"syscall"
)

type db struct {
	data []byte
}

func (db *db) set(key string, value string) {
	fmt.Println("DB before modification: ", string(db.data))
	s := key + ": " + value
	b := []byte(s)
	for i := 0; i < len(b); i++ {
		db.data[i] = b[i]
	}
	fmt.Println("DB after modification: ", string(db.data))
}

func (db *db) handler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	switch r.Method {
	case "GET":
		fmt.Print(ps.ByName("key"))
		fmt.Fprintf(w, "Getting %s!", ps.ByName("key"))
	case "POST":
		db.set(ps.ByName("key"), r.FormValue("value"))
	}
}

func NewHandler(db *db) func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	return db.handler
}

func openDB() *db {
	filename := "db"

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		os.Create("db")
	}

	f, err := os.OpenFile(filename, os.O_RDWR, 0)
	if err != nil {
		fmt.Println("Could not open file: ", err)
	}

	data, err := syscall.Mmap(int(f.Fd()), 0, 100, syscall.PROT_WRITE|syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		fmt.Printf("Could not memory map the file (into a byte array): ", err)
	}

	m := &db{data}
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
