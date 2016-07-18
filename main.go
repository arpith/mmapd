package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"os"
	"strings"
)

func (db *db) handler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	switch r.Method {
	case "GET":
		key := ps.ByName("key")
		c := make(chan string)
		m := readChanMessage{key, c}
		db.readChan <- m
		m = <-c
		close(c)
		if m.err == "Invalid Key" {
			http.NotFound(w, r)
		} else {
			fmt.Fprint(w, m.json)
		}
	case "POST":
		m := make(map[string]string)
		m["key"] = ps.ByName("key")
		m["value"] = r.FormValue("value")
		db.writeChan <- m
		fmt.Fprint(w, r.FormValue("value"))
	}
}

func NewHandler(db *db) func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	return db.handler
}

func main() {
	dbFilename := "../db.json"
	logFilename := "../log.json"
	db := initDB(dbFilename, logFilename)
	handler := NewHandler(db)

	router := httprouter.New()
	router.GET("/get/:key", handler)
	router.POST("/set/:key", handler)

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "3001"
	}

	http.ListenAndServe(":"+port, router)
}
