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
	fmt.Println("Haven't figured this out yet!")
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
