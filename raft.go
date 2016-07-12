package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"syscall"
)

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
	ip               string
	state            string
	term             term
	electionTimeout  int
	heartbeatTimeout int
	config           []string
	receiveChan      chan string
}

func (server *server) sendRequestForVotes(server) {
	form := url.Values{"id": {server.ip}, "term": {server.term.id}}
	resp, err := http.PostForm(server+"/votes", form)
	if err != nil {
		fmt.Println("Couldn't send request for votes to " + server)
	}
	defer resp.Body.Close()
}

func (server *server) startElectionTerm() {
	server.state = "candidate"
	server.term += 1
	server.term.votes += 1
	for _, server := range server.config {
		go server.sendRequestForVotes(server)
	}
}

func (server *server) sendAppendEntryRequests() {
}

func (server *server) stepDown() {

}

func (server *server) handleVote() {
}

func (server *server) handleRequestForVote(equest voteRequest, w httpResponse) {
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

func (server *server) handleAppendEntryRequest(request appendEntryRequest) {
	if request.termID < server.term.id {
		fmt.Fprint(w, false)
	} else if server.log[request.prevLogIndex].term != req.prevLogTerm {
		fmt.Fprint(w, false)
	} else if server.log[req.index].term != req.term {
		server.log = server.log[:req.index]
	}
	server.log = server.log.appendEntry(req.entry)
	if req.commitIndex > server.commitIndex {
		s.commitIndex = Min(req.commitIndex, req.entry.index)
	}

}

func (server *server) appendEntry(entry entry) {
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

func initServer(ip string) *server {
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
