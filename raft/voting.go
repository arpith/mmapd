package raft

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type requestForVote struct {
	Term         int
	CandidateID  string
	LastLogIndex int
	LastLogTerm  int
}

type voteRequest struct {
	Req        requestForVote
	ReturnChan chan requestForVoteResponse
}

type requestForVoteResponse struct {
	Term           int
	HasGrantedVote bool
}

type voteResponse struct {
	ServerIndex int
	Resp        requestForVoteResponse
}

func (s *server) handleRequestForVote(v voteRequest) {
	req := v.Req
	returnChan := v.ReturnChan
	if req.Term < s.term {
		resp := &requestForVoteResponse{s.term, false}
		returnChan <- *resp
	} else {
		cond1 := s.votedFor == ""
		cond2 := s.votedFor == req.CandidateID
		cond3 := req.LastLogIndex >= len(s.db.Log.Entries)
		if (cond1 || cond2) && cond3 {
			s.electionTimeout.reset()
			s.votedFor = req.CandidateID
			s.term = req.Term
			resp := &requestForVoteResponse{s.term, true}
			returnChan <- *resp
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
	v := &requestForVote{s.term, s.id, lastLogIndex, lastLogTerm}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(v)
	resp, err := http.Post("http://"+receiver+"/votes", "application/json", b)
	if err != nil {
		fmt.Println("Couldn't send request for votes to " + receiver)
		fmt.Println(err)
		return
	} else {
		r := &requestForVoteResponse{}
		err := json.NewDecoder(resp.Body).Decode(r)
		if err != nil {
			fmt.Println("Couldn't decode request for vote response from ", s.config[receiverIndex])
			return
		}
		defer resp.Body.Close()
		voteResp := &voteResponse{receiverIndex, *r}
		respChan <- *voteResp
	}
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
			if vote.Resp.HasGrantedVote {
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
