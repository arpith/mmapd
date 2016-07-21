package raft

import (
	"encoding/json"
	"fmt"
	"github.com/arpith/mmapd/db"
	"net/http"
	"net/url"
	"strconv"
)

type appendEntryRequest struct {
	term         int
	leaderId     string
	prevLogIndex int
	prevLogTerm  int
	entry        db.Entry
	leaderCommit int
	returnChan   chan bool
}

type appendEntryResponse struct {
	term    int
	success bool
}

type followerResponse struct {
	serverIndex int
	resp        appendEntryResponse
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
	v := url.Values{}
	v.Set("term", strconv.Itoa(s.term))
	v.Set("leaderID", s.id)
	v.Set("prevLogIndex", strconv.Itoa(len(s.db.Log.Entries)))
	v.Set("entry", entry.Command)
	v.Set("leaderCommit", strconv.Itoa(s.commitIndex))
	resp, err := http.PostForm(follower+"/appendEntry", v)
	if err != nil {
		fmt.Println("Couldn't send append entry request to " + follower)
		go s.sendAppendEntryRequest(followerIndex, entry, respChan)
		return
	}
	r := &appendEntryResponse{}
	json.NewDecoder(resp.Body).Decode(r)
	defer resp.Body.Close()
	if r.term > s.term {
		s.term = r.term
		s.state = "follower"
		s.electionTimeout.reset()
	}
	if r.success {
		s.term = r.term
		s.nextIndex[followerIndex]++
		s.matchIndex[followerIndex]++
		followerResp := &followerResponse{followerIndex, *r}
		respChan <- *followerResp
	} else {
		s.nextIndex[followerIndex]--
		go s.sendAppendEntryRequest(followerIndex, entry, respChan)
	}
}

func (s *server) handleAppendEntryRequest(req appendEntryRequest) {
	if req.term < s.term {
		req.returnChan <- false
	} else if s.db.Log.Entries[req.prevLogIndex].Term != req.prevLogTerm {
		req.returnChan <- false
	} else {
		if s.db.Log.Entries[req.prevLogIndex+1].Term != req.term {
			// If existing entry conflicts with new entry
			// Delete entry and all that follow it
			s.db.Log.SetEntries(s.db.Log.Entries[:req.prevLogIndex])
		}
		s.db.Log.AppendEntry(req.entry)
		if req.leaderCommit < s.commitIndex {
			// Set commit index to the min of the leader's commit index and index of last new entry
			if req.leaderCommit < req.prevLogIndex+1 {
				s.commitIndex = req.leaderCommit
			} else {
				s.commitIndex = req.prevLogIndex + 1
			}
		}
		req.returnChan <- true
	}
}
