package raft

import (
	"encoding/json"
	"fmt"
	"github.com/arpith/mmapd/db"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (s *server) appendEntryHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var a appendEntryRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&a)
	if err != nil {
		fmt.Println("Couldn't decode append entry request as json", err)
	}
	returnChan := make(chan appendEntryResponse)
	req := &appendRequest{
		Req:        a,
		ReturnChan: returnChan,
	}
	s.appendRequests <- *req
	resp := <-req.ReturnChan
	defer close(req.ReturnChan)
	json.NewEncoder(w).Encode(resp)
}

func (s *server) requestForVoteHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var v requestForVote
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&v)
	if err != nil {
		fmt.Println("Couldn't decode vote request as json", err)
	}
	returnChan := make(chan requestForVoteResponse)
	req := &voteRequest{
		Req:        v,
		ReturnChan: returnChan,
	}
	s.voteRequests <- *req
	resp := <-req.ReturnChan
	defer close(req.ReturnChan)
	json.NewEncoder(w).Encode(resp)
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

func (s *server) statusRequestHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	status := &status{
		s.id,
		s.state,
		s.term,
		s.votedFor,
		s.commitIndex,
		s.lastApplied,
		s.nextIndex,
		s.matchIndex,
	}
	fmt.Println(status)
	json.NewEncoder(w).Encode(*status)
}

func NewHandler(s *server, handlerType string) func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	switch handlerType {
	case "Append Entry":
		return s.appendEntryHandler
	case "Request For Vote":
		return s.requestForVoteHandler
	case "Status Request":
		return s.statusRequestHandler
	default:
		return s.clientRequestHandler
	}
}
