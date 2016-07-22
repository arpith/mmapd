package raft

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
			s.electionTimeout.reset()
			s.votedFor = req.candidateId
			s.term = req.term
			req.returnChan <- true
		}
	}
}

func (s *server) sendRequestForVote(receiverIndex int, respChan chan voteResponse) {
	receiver := s.config[receiverIndex]
	lastLogIndex := len(s.db.Log.Entries)
	lastLogTerm := 0
	if lastLogIndex > 0 {
		lastLogTerm = s.db.Log.Entries[lastLogIndex-1].Term
	}
	v := url.Values{}
	v.Set("candidateID", s.id)
	v.Set("term", strconv.Itoa(s.term))
	v.Set("lastLogIndex", strconv.Itoa(lastLogIndex))
	v.Set("lastLogTerm", strconv.Itoa(lastLogTerm))
	resp, err := http.PostForm("http://"+receiver+"/votes", v)
	if err != nil {
		fmt.Println("Couldn't send request for votes to " + receiver)
	}
	r := &requestForVoteResponse{}
	json.NewDecoder(resp.Body).Decode(r)
	defer resp.Body.Close()
	voteResp := &voteResponse{receiverIndex, *r}
	respChan <- *voteResp
}

func (s *server) startElection() {
	s.state = "candidate"
	s.term += 1
	voteCount := 1
	respChan := make(chan voteResponse)
	for receiverIndex, receiverId := range s.config {
		if receiverId != s.id {
			go s.sendRequestForVote(receiverIndex, respChan)
		}
	}
	responseCount := 0
	if len(s.config) > 1 {
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
}
