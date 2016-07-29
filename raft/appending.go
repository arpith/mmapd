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

func (s *server) appendEntry(command string, key string, value string, isCommitted chan bool) {
	entry := &db.Entry{command, key, value, s.term}
	index := -1
	if command != "" {
		s.db.Log.AppendEntry(*entry)
		index = len(s.db.Log.Entries) - 1
	}
	respChan := make(chan followerResponse)
	for i := 0; i < len(s.config); i++ {
		if s.config[i] != s.id {
			go s.sendAppendEntryRequest(i, index, respChan)
		} else {
			if command != "" {
				//Update nextIndex and matchIndex for self
				s.nextIndex[i]++
				s.matchIndex[i]++
			}
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
				return
			}
			if responseCount == len(s.config)-1 {
				fmt.Println("Got all responses!")
				/*
						isCommitted <- false
					return
				*/
			}
		}
	} else {
		if command != "" {
			s.commitIndex++
			isCommitted <- true
		}
	}
}

func (s *server) sendAppendEntryRequest(followerIndex int, entryIndex int, respChan chan followerResponse) {
	entryP := &db.Entry{"", "", "", s.term}
	entry := *entryP
	follower := s.config[followerIndex]
	prevLogIndex := -1
	prevLogTerm := 0
	if entryIndex == -1 {
		fmt.Println(s.nextIndex)
	}
	if entryIndex == -1 && len(s.db.Log.Entries)-1 > s.nextIndex[followerIndex] {
		// When sending a heartbeat, if nextIndex < last log index, send the missing entry!
		entryIndex = s.nextIndex[followerIndex]
	}
	if entryIndex > -1 {
		if entryIndex < len(s.db.Log.Entries) {
			entry = s.db.Log.Entries[entryIndex]
		}
		prevLogIndex = entryIndex - 1
	}
	if prevLogIndex >= 0 && len(s.db.Log.Entries) > prevLogIndex {
		prevLogTerm = s.db.Log.Entries[prevLogIndex].Term
	}
	fmt.Println("Going to send Append Entry RPC to ", follower, " for entry ", entry, " (prevLogIndex: ", prevLogIndex, " )")
	a := &appendEntryRequest{s.term, s.id, prevLogIndex, prevLogTerm, entry, s.commitIndex}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(a)
	resp, err := http.Post("http://"+follower+"/append", "application/json", b)
	if err != nil {
		fmt.Println("Couldn't send append entry request to " + follower)
		fmt.Println(err)
		r := &appendEntryResponse{Success: false}
		f := &followerResponse{followerIndex, *r}
		respChan <- *f
		//go s.sendAppendEntryRequest(followerIndex, entryIndex, respChan)
	} else {
		r := &appendEntryResponse{}
		err := json.NewDecoder(resp.Body).Decode(r)
		if err != nil {
			fmt.Println("Couldn't decode append entries response from " + s.config[followerIndex])
			return
		}
		resp.Body.Close()
		if r.Term > s.term {
			s.term = r.Term
			s.stepDown("Got append entry RPC response with term > current term")
		}
		if r.Success {
			s.term = r.Term
			if entryIndex != -1 {
				s.nextIndex[followerIndex]++
				s.matchIndex[followerIndex]++
			}
			followerResp := &followerResponse{followerIndex, *r}
			respChan <- *followerResp
		} else {
			if s.nextIndex[followerIndex] > -1 {
				s.nextIndex[followerIndex]--
			}
			fmt.Println(s.nextIndex[followerIndex], len(s.db.Log.Entries))
			go s.sendAppendEntryRequest(followerIndex, s.nextIndex[followerIndex], respChan)
		}
	}
}

func (s *server) handleAppendEntryRequest(a appendRequest) {
	fmt.Println("Append entry: ", a.Req)
	returnChan := a.ReturnChan
	req := a.Req
	if req.Term < s.term {
		resp := &appendEntryResponse{s.term, false}
		returnChan <- *resp
	} else if req.PrevLogIndex > -1 &&
		len(s.db.Log.Entries) > req.PrevLogIndex &&
		s.db.Log.Entries[req.PrevLogIndex].Term != req.PrevLogTerm {
		resp := &appendEntryResponse{s.term, false}
		returnChan <- *resp
	} else {
		if req.Term > s.term {
			s.term = req.Term
			s.stepDown("append entries RPC has term greater than current term")
		}
		if req.Entry.Command == "" {
			s.term = req.Term
			s.stepDown("append entries RPC has term >= current term AND is a heartbeat")
			fmt.Println("got a heartbeat, responding true: term >= current term")
		} else {
			if len(s.db.Log.Entries) > req.PrevLogIndex+2 {
				if s.db.Log.Entries[req.PrevLogIndex+1].Term != req.Term {
					// If existing entry conflicts with new entry
					// Delete entry and all that follow it
					s.db.Log.SetEntries(s.db.Log.Entries[:req.PrevLogIndex+1])
				}
			}
			s.db.Log.AppendEntry(req.Entry)
		}
		if req.LeaderCommit > s.commitIndex {
			// Set commit index to the min of the leader's commit index and index of last new entry
			if req.Entry.Command == "" || req.LeaderCommit < req.PrevLogIndex+1 {
				fmt.Println("Going to commit entries with leader commit: ", req.LeaderCommit)
				s.commitEntries(req.LeaderCommit)
			} else {
				fmt.Println("Going to commit entries with prevLogIndex: ", req.PrevLogIndex)
				s.commitEntries(req.PrevLogIndex + 1)
			}
		}
		resp := &appendEntryResponse{s.term, true}
		returnChan <- *resp
	}
}
