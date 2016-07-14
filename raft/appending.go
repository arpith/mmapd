package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"syscall"
)

type appendEntriesResponse struct {
	term    int
	success bool
}

func (s *server) sendAppendEntryRequest(followerIndex, entry) {
	follower := s.config[followerIndex]
	v := url.Values{}
	v.set("term", s.term)
	v.set("leaderID", s.id)
	v.set("prevLogIndex", len(s.db.log))
	v.set("entry", entry)
	v.set("leaderCommit", s.commitIndex)
	resp, err := http.PostForm(follower+"/appendEntry", v)
	if err != nil {
		fmt.Println("Couldn't send append entry request to " + follower)
	}
	r := &AppendEntriesResponse{}
	json.NewDecoder(resp.Body).Decode(r)
	if r.term > s.term {
		s.term = r.term
		s.convertToFollower()
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
	for N := s.commitIndex + 1; N < len(s.db.log); N++ {
		//Check if there exists an N > commitIndex
		count := 0
		for i := 0; i < len(s.matchIndex); i++ {
			if s.matchIndex[i] >= N {
				count++
			}
			// Check if a majority of matchIndex[i] >= N
			cond1 := count > len(s.matchIndex)/2
			// Check if log[N].term == currentTerm
			cond2 := s.db.log[N].term == s.term
			if cond1 && cond2 {
				// Set commitIndex to N
				s.commitIndex = N
				break
			}
		}
	}
}

func (s *server) handleAppendEntryRequest(req appendEntryRequest, w http.ResponseWriter) {
	if req.termID < s.term.id {
		fmt.Fprint(w, false)
	} else if s.db.log[request.prevLogIndex].term != req.prevLogTerm {
		fmt.Fprint(w, false)
	} else if s.db.log[req.index].term != req.term {
		s.db.log = s.db.log[:req.index]
	}
	s.db.log = s.db.log.appendEntry(req.entry)
	if req.commitIndex > s.commitIndex {
		s.commitIndex = Min(req.commitIndex, req.entry.index)
	}
	fmt.Fprint(w, true)
}
