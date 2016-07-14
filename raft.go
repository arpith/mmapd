package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"syscall"
)

type appendEntriesResponse struct {
	term    string
	success bool
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

type server struct {
	id               string
	term             string
	db               *db
	electionTimeout  int
	heartbeatTimeout int
	config           []string
	receiveChan      chan string
	commitIndex      int
	lastApplied      int
	nextIndex        []int
	matchIndex       []int
}

func (s *server) sendRequestForVotes(server) {
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

func (server *server) startElectionTerm() {
	server.state = "candidate"
	server.term += 1
	server.term.votes += 1
	for _, server := range server.config {
		go server.sendRequestForVotes(server)
	}
}

func (server *server) stepDown() {

}

func (server *server) handleVote() {
}

func (server *server) handleRequestForVote(request voteRequest, w http.ResponseWriter) {
	if request.termID < server.term.id {
		fmt.Fprint(w, false)
	} else {
		cond1 := server.term.vote == ""
		cond2 := server.term.vote == request.candidateID
		cond3 := request.lastLogIndex >= server.lastLogIndex
		if (cond1 || cond2) && cond3 {
			server.term.vote = request.candidateID
			server.term.id = request.termID
			fmt.Fprint(w, true)
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

func (server *server) listener() {
	select {
	case v := <-server.requestForVote:
		server.handleRequestForVote(v)
	case e := <-server.appendEntry:
		server.handleAppendEntryRequest(e)
	case <-time.After(server.electionTimeout * time.Millisecond):
		server.startElectionTerm()
	case <-time.After(server.heartbeatTimeout * time.Millisecond):
		server.appendEntry("")
	}
}

func readConfig(filename) []string {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Couldn't read config file")
	}
	servers := strings.Split(string(content), "\n")
	return servers
}

func initServer(ip string, db *db) *server {
	state := "follower"
	term := &term{0, false, 0}
	electionTimeout := 150 + rand.Int(rand.Reader, 150)
	heartbeatTimeout := 150 + rand.Int(rand.Reader, 150)
	config = readConfig("config.txt")
	receiveChan := make(chan string)
	server := &server{ip, state, term, electionTimeout, heartbeatTimeout, config, receiveChan}
	go server.listener()
	return server
}
