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

func (s *server) stepDown(reason string) {
	fmt.Println("SETTING FOLLOWER & RESETTING ELECTION TIMEOUT: ", reason)
	s.state = "follower"
	s.electionTimeout.reset()
}

func (s *server) becomeLeader() {
	//Initialize nextIndex values to the index after the last one in leader's log
	for follower, _ := range s.nextIndex {
		s.nextIndex[follower] = len(s.db.Log.Entries)
	}
	fmt.Println(s.nextIndex)
	fmt.Println("SETTING LEADER: got majority vote")
	s.state = "leader"
	fmt.Println("IM THE LEADER :D :D :D :D :D ")
}

func (s *server) handleRequestForVote(v voteRequest) {
	fmt.Println("Got vote request:", v.Req)
	req := v.Req
	returnChan := v.ReturnChan
	if req.Term < s.term {
		resp := &requestForVoteResponse{s.term, false}
		returnChan <- *resp
	} else {
		if req.Term > s.term {
			s.votedFor = ""
			s.stepDown("Got vote request with term > current term")
		}
		cond1 := s.votedFor == ""
		cond2 := s.votedFor == req.CandidateID
		cond3 := req.LastLogIndex >= len(s.db.Log.Entries)-1
		if (cond1 || cond2) && cond3 {
			s.votedFor = req.CandidateID
			s.term = req.Term
			resp := &requestForVoteResponse{s.term, true}
			returnChan <- *resp
		}
	}
}

func (s *server) sendRequestForVote(receiverIndex int, respChan chan voteResponse) {
	receiver := s.config[receiverIndex]
	lastLogIndex := len(s.db.Log.Entries) - 1
	lastLogTerm := 0
	if lastLogIndex > 0 {
		lastLogTerm = s.db.Log.Entries[lastLogIndex].Term
	}
	v := &requestForVote{s.term, s.id, lastLogIndex, lastLogTerm}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(v)
	resp, err := http.Post("http://"+receiver+"/votes", "application/json", b)
	if err != nil {
		r := &requestForVoteResponse{0, false}
		v := &voteResponse{receiverIndex, *r}
		fmt.Println("Couldn't send request for votes to " + receiver)
		fmt.Println(err)
		respChan <- *v
	} else {
		r := &requestForVoteResponse{}
		err := json.NewDecoder(resp.Body).Decode(r)
		if err != nil {
			r := &requestForVoteResponse{0, false}
			v := &voteResponse{receiverIndex, *r}
			fmt.Println("Couldn't decode request for vote response from ", s.config[receiverIndex])
			respChan <- *v
			return
		}
		resp.Body.Close()
		voteResp := &voteResponse{receiverIndex, *r}
		respChan <- *voteResp
	}
}

func (s *server) startElection() {
	fmt.Println("SETTING CANDIDATE: Going to start election")
	s.state = "candidate"
	s.term += 1
	s.votedFor = s.id
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
			fmt.Println("Got vote response", vote)
			if vote.Resp.Term > s.term {
				s.stepDown("Got vote response with term greater than current term")
				break
			}
			if vote.Resp.HasGrantedVote {
				voteCount++
			}
			if voteCount > (len(s.config)-1)/2 {
				s.becomeLeader()
				break
			}
			if responseCount == len(s.config)-2 {
				fmt.Println("Got all the responses")
				break
			}
		}
	}
}
