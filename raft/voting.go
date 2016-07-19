package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type voteRequest struct {
	term         int
	candidateId  string
	lastLogIndex int
	lastLogTerm  int
	returnChan   chan bool
}

type requestForVoteResponse struct {
	term           int
	hasGrantedVote bool
}

type voteResponse struct {
	serverIndex int
	resp        requestForVoteResponse
}

func (s *server) handleRequestForVote(req voteRequest) {
	if req.term < s.term {
		req.returnChan <- false
	} else {
		cond1 := s.votedFor == ""
		cond2 := s.votedFor == req.candidateId
		cond3 := req.lastLogIndex >= len(s.db.Log.Entries)
		if (cond1 || cond2) && cond3 {
			s.electionTimeout.resetTimeout()
			s.votedFor = req.candidateId
			s.term = req.term
			req.returnChan <- true
		}
	}
}

func (s *server) sendRequestForVote(receiverIndex int, respChan chan voteResponse) {
	receiver := s.config[receiverIndex]
	lastLogIndex := len(s.db.Log.Entries)
	lastLogTerm := s.db.Log.Entries[lastLogIndex-1].Term
	v := url.Values{}
	v.Set("candidateID", s.id)
	v.Set("term", strconv.Itoa(s.term))
	v.Set("lastLogIndex", strconv.Itoa(lastLogIndex))
	v.Set("lastLogTerm", strconv.Itoa(lastLogTerm))
	resp, err := http.PostForm(receiver+"/votes", v)
	if err != nil {
		fmt.Println("Couldn't send request for votes to " + receiver)
	}
	defer resp.Body.Close()
	r := &requestForVoteResponse{}
	json.NewDecoder(resp.Body).Decode(r)
	voteResp := &voteResponse{receiverIndex, *r}
	respChan <- *voteResp
}

func (s *server) startElection() {
	s.state = "candidate"
	s.term += 1
	voteCount := 1
	respChan := make(chan voteResponse)
	for receiverIndex, _ := range s.config {
		go s.sendRequestForVote(receiverIndex, respChan)
	}
	responseCount := 0
	for {
		vote := <-respChan
		responseCount++
		if vote.resp.hasGrantedVote {
			voteCount++
		}
		if voteCount > len(s.config)/2 {
			s.state = "leader"
			break
		}
		if responseCount == len(s.config) {
			break
		}
	}
}
