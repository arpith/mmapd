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

func (s *server) appendEntry(entry string) {
	for i := 0; i < len(s.config); i++ {
		if s.config[i] != s.id {
			go s.sendAppendEntryRequest(i, entry)
		}
	}
}

func (s *server) sendAppendEntryRequest(followerIndex int, entry string) {
	follower := s.config[followerIndex]
	v := url.Values{}
	v.Set("term", strconv.Itoa(s.term))
	v.Set("leaderID", s.id)
	v.Set("prevLogIndex", strconv.Itoa(len(s.db.Log.Entries)))
	v.Set("entry", entry)
	v.Set("leaderCommit", strconv.Itoa(s.commitIndex))
	resp, err := http.PostForm(follower+"/appendEntry", v)
	if err != nil {
		fmt.Println("Couldn't send append entry request to " + follower)
	}
	r := &appendEntryResponse{}
	json.NewDecoder(resp.Body).Decode(r)
	if r.term > s.term {
		s.term = r.term
		s.state = "follower"
		s.electionTimeout.reset()
	}
	if r.success {
		s.term = r.term
		s.nextIndex[followerIndex]++
		s.matchIndex[followerIndex]++
	} else {
		s.nextIndex[followerIndex]--
		go s.sendAppendEntryRequest(followerIndex, entry)
	}
	defer resp.Body.Close()
	for N := s.commitIndex + 1; N < len(s.db.Log.Entries); N++ {
		//Check if there exists an N > commitIndex
		count := 0
		for i := 0; i < len(s.matchIndex); i++ {
			if s.matchIndex[i] >= N {
				count++
			}
			// Check if a majority of matchIndex[i] >= N
			cond1 := count > len(s.matchIndex)/2
			// Check if log[N].term == currentTerm
			cond2 := s.db.Log.Entries[N].Term == s.term
			if cond1 && cond2 {
				// Set commitIndex to N
				s.commitIndex = N
				break
			}
		}
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
