package db

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (db *DB) handler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	switch r.Method {
	case "GET":
		key := ps.ByName("key")
		c := make(chan returnChanMessage)
		m := readChanMessage{key, c}
		db.readChan <- m
		resp := <-c
		close(c)
		if resp.err != nil {
			http.NotFound(w, r)
		} else {
			fmt.Fprint(w, resp.json)
		}
	case "POST":
		m := make(map[string]string)
		m["key"] = ps.ByName("key")
		m["value"] = r.FormValue("value")
		db.writeChan <- m
		fmt.Fprint(w, r.FormValue("value"))
	}
}

func NewHandler(db *DB) func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	return db.handler
}
