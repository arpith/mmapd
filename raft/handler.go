package raft

import (
	"encoding/json"
	"fmt"
	"github.com/arpith/mmapd/db"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (s *server) appendEntryHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	returnChan := make(chan appendEntryResponse)
	req := &appendEntryRequest{
		Term:         strconv.r.FormValue("term"),
		LeaderID:     r.FormValue("leaderID"),
		PrevLogIndex: r.FormValue("prevLogIndex"),
		Entry:        r.FormValue("entry"),
		LeaderCommit: r.FormValue("leaderCommit"),
		ReturnChan:   returnChan,
	}
	s.appendEntryRequests <- req
	resp := <-req.ReturnChan
	defer close(req.ReturnChan)
	json.NewEncoder(w).Encode(resp)
}

func (s *server) requestForVoteHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	returnChan := make(chan appendEntryResponse)
	req := &voteRequest{
		Term:         r.FormValue("term"),
		CandidateID:  r.FormValue("cadidateID"),
		LastLogIndex: r.FormValue("lastLogIndex"),
		LastLogTerm:  r.FormValue("lastLogTerm"),
		ReturnChan:   returnChan,
	}
	s.appendEntryRequests <- req

	s.voteRequests <- req
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
