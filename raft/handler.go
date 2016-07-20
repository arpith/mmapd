package raft

import (
	"encoding/json"
	"fmt"
	"github.com/arpith/mmapd/db"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (s *server) appendEntryHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	decoder := json.NewDecoder(r.Body)
	var req appendEntryRequest
	err := decoder.Decode(&req)
	if err != nil {
		fmt.Println("Error decoding request: ", err)
	}
	s.appendEntryRequests <- req
}

func (s *server) requestForVoteHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	decoder := json.NewDecoder(r.Body)
	var req voteRequest
	err := decoder.Decode(&req)
	if err != nil {
		fmt.Println("Error decoding request: ", err)
	}
	s.voteRequests <- req
}

func (s *server) clientRequestHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	switch r.Method {
	case "GET":
		key := ps.ByName("key")
		c := make(chan db.ReturnChanMessage)
		m := readRequest{key, c}
		s.readRequests <- m
		resp := <-c
		close(c)
		if resp.Err != nil {
			http.NotFound(w, r)
		} else {
			fmt.Fprint(w, resp.Json)
		}
	case "POST":
		key := ps.ByName("key")
		value := r.FormValue("value")
		c := make(chan db.ReturnChanMessage)
		m := writeRequest{key, value, c}
		s.writeRequests <- m
		resp := <-c
		close(c)
		if resp.Err != nil {
			fmt.Fprint(w, resp.Err)
		} else {
			fmt.Fprint(w, resp.Json)
		}
	}
}

func NewHandler(s *server, handlerType string) func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	switch handlerType {
	case "Append Entry":
		return s.appendEntryHandler
	case "Request For Vote":
		return s.requestForVoteHandler
	default:
		return s.clientRequestHandler
	}
}
