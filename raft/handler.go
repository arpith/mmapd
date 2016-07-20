package raft

import (
	"encoding/json"
	"fmt"
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
		c := make(chan returnChanMessage)
		m := readRequest{key, c}
		s.readRequests <- m
		resp := <-c
		close(c)
		if resp.err != nil {
			http.NotFound(w, r)
		} else {
			fmt.Fprint(w, resp.json)
		}
	case "POST":
		key := ps.ByName("key")
		value := r.FormValue("value")
		c := make(chan returnChanMessage)
		m := writeRequest{key, value, c}
		s.writeRequests <- m
		resp := <-c
		close(c)
		if resp.err != nil {
			fmt.Fprint(w, resp.err)
		} else {
			fmt.Fprint(w, resp.json)
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
