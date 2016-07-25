package raft

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/arpith/mmapd/db"
	"net/http"
)

type appendEntryRequest struct {
	Term         int
	LeaderID     string
	PrevLogIndex int
	PrevLogTerm  int
	Entry        db.Entry
	LeaderCommit int
}

type appendRequest struct {
	Req        appendEntryRequest
	ReturnChan chan appendEntryResponse
}

type appendEntryResponse struct {
	Term    int
	Success bool
}

type followerResponse struct {
	ServerIndex int
	Resp        appendEntryResponse
}

func (s *server) appendEntry(command string, isCommitted chan bool) {
	entry := &db.Entry{command, s.term}
	index := -1
	if command != "" {
		s.db.Log.AppendEntry(*entry)
		index = len(s.db.Log.Entries)
	}
	respChan := make(chan followerResponse)
	for i := 0; i < len(s.config); i++ {
		if s.config[i] != s.id {
			go s.sendAppendEntryRequest(i, *entry, respChan)
		}
	}
	responseCount := 0
	if len(s.config) > 1 {
		for {
			_ = <-respChan
			responseCount++
			for N := s.commitIndex + 1; N < len(s.db.Log.Entries); N++ {
				//Check if there exists an N > commitIndex
				count := 0
				for i := 0; i < len(s.matchIndex); i++ {
					if s.matchIndex[i] >= N {
						count++
					}
				}
				// Check if a majority of matchIndex[i] >= N
				cond1 := count > len(s.matchIndex)/2
				// Check if log[N].term == currentTerm
				cond2 := s.db.Log.Entries[N].Term == s.term
				if cond1 && cond2 {
					// Set commitIndex to N
					s.commitIndex = N
				} else {
					break
				}
			}
			if s.commitIndex == index {
				isCommitted <- true
			}
		}
	} else {
		if command != "" {
			s.commitIndex++
			isCommitted <- true
		}
	}
}

func (s *server) sendAppendEntryRequest(followerIndex int, entry db.Entry, respChan chan followerResponse) {
	follower := s.config[followerIndex]
	prevLogIndex := len(s.db.Log.Entries) - 1
	prevLogTerm := s.db.Log.Entries[prevLogIndex-1].Term
	a := &appendEntryRequest{s.term, s.id, prevLogIndex, prevLogTerm, entry, s.commitIndex}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(a)
	resp, err := http.Post("http://"+follower+"/append", "application/json", b)
	if err != nil {
		fmt.Println("Couldn't send append entry request to " + follower)
		fmt.Println(err)
		return
		//		go s.sendAppendEntryRequest(followerIndex, entry, respChan)
	} else {
		r := &appendEntryResponse{}
		err := json.NewDecoder(resp.Body).Decode(r)
		if err != nil {
			fmt.Println("Couldn't decode append entries response from " + s.config[followerIndex])
			return
		}
		defer resp.Body.Close()
		if r.Term > s.term {
			s.term = r.Term
			s.state = "follower"
			s.electionTimeout.reset()
		}
		if r.Success {
			s.term = r.Term
			s.nextIndex[followerIndex]++
			s.matchIndex[followerIndex]++
			followerResp := &followerResponse{followerIndex, *r}
			respChan <- *followerResp
		} else {
			s.nextIndex[followerIndex]--
			//		go s.sendAppendEntryRequest(followerIndex, entry, respChan)
		}
	}
}

func (s *server) handleAppendEntryRequest(a appendRequest) {
	returnChan := a.ReturnChan
	req := a.Req
	if req.Term < s.term {
		resp := &appendEntryResponse{s.term, false}
		returnChan <- *resp
	} else if len(s.db.Log.Entries) > req.PrevLogIndex && s.db.Log.Entries[req.PrevLogIndex].Term != req.PrevLogTerm {
		resp := &appendEntryResponse{s.term, false}
		returnChan <- *resp
	} else {
		if s.db.Log.Entries[req.PrevLogIndex+1].Term != req.Term {
			// If existing entry conflicts with new entry
			// Delete entry and all that follow it
			s.db.Log.SetEntries(s.db.Log.Entries[:req.PrevLogIndex])
		}
		s.db.Log.AppendEntry(req.Entry)
		if req.LeaderCommit < s.commitIndex {
			// Set commit index to the min of the leader's commit index and index of last new entry
			if req.LeaderCommit < req.PrevLogIndex+1 {
				s.commitIndex = req.LeaderCommit
			} else {
				s.commitIndex = req.PrevLogIndex + 1
			}
		}
		resp := &appendEntryResponse{s.term, true}
		returnChan <- *resp
	}
}
