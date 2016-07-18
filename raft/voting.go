package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type voteRequest struct {
	term         int
	candidateId  string
	lastLogIndex int
	lastLogTerm  int
}

type requestForVotesResponse struct {
	term           int
	hasGrantedVote bool
}

type voteResponse struct {
	serverIndex int
	resp        requestForVotesResponse
}

func (s *server) handleRequestForVote(request voteRequest, w http.ResponseWriter) {
	if request.term < s.term {
		fmt.Fprint(w, false)
	} else {
		cond1 := s.votedFor == ""
		cond2 := s.votedFor == request.candidateID
		cond3 := request.lastLogIndex >= len(s.log)
		if (cond1 || cond2) && cond3 {
			s.electionTimeout.resetTimeout()
			s.votedFor = request.candidateID
			s.term = request.term
			fmt.Fprint(w, true)
		}
	}
}

func (s *server) sendRequestForVote(receiver string, respChan chan RequestForVoteResponse) {
	v := url.Values{}
	v.set("candidateID", s.id)
	v.set("term", s.term)
	v.set("lastLogIndex", len(s.db.log))
	v.set("lastLogTerm", s.db.log[len(s.db.log)-1].term)
	resp, err := http.PostForm(server+"/votes", v)
	if err != nil {
		fmt.Println("Couldn't send request for votes to " + server)
	}
	defer resp.Body.Close()
	r = &RequestForVoteResponse{}
	json.NewDecoder(resp.Body).Decode(r)
	v = &VoteResponse{receiver, r}
	respChan <- v
}

func (server *server) startElection() {
	server.state = "candidate"
	server.term += 1
	server.term.votes += 1
	respChan = make(chan voteResponse)
	for receiverIndex, _ := range server.config {
		go server.sendRequestForVotes(receiverIndex)
	}
	voteCount = 0
	responseCount = 0
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
