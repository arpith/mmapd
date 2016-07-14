package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"syscall"
)

type requestForVotesResponse struct {
	term           int
	hasGrantedVote bool
}

type voteResponse struct {
	serverIndex int
	resp        requestForVotesResponse
}

type voteRequest struct {
	ip     string
	termID int
}

type term struct {
	id       int
	votes    int
	votedFor string
}

func (s *server) sendRequestForVotes(receiver string, respChan chan RequestForVotesResponse) {
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
	r = &RequestForVotesResponse{}
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
	for {
		vote := <-respChan
		if voteResponse.resp.hasGrantedVote {
			voteCount++
		}
		if voteCount > len(s.config)/2 {
			s.state = "leader"
			break
		}
	}

}

func (s *server) handleRequestForVote(request voteRequest, w http.ResponseWriter) {
	if request.termID < s.term.id {
		fmt.Fprint(w, false)
	} else {
		cond1 := s.term.vote == ""
		cond2 := s.term.vote == request.candidateID
		cond3 := request.lastLogIndex >= s.lastLogIndex
		if (cond1 || cond2) && cond3 {
			s.electionTimeout.resetTimeout()
			s.term.vote = request.candidateID
			s.term.id = request.termID
			fmt.Fprint(w, true)
		}
	}
}
