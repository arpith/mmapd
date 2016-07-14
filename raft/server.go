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

type appendEntriesResponse struct {
	term    int
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
	term             int
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

func (server *server) listener() {
	for {
		select {
		case v := <-server.requestForVote:
			server.handleRequestForVote(v)
		case e := <-server.appendEntry:
			server.handleAppendEntryRequest(e)
		case <-time.After(server.electionTimeout * time.Millisecond):
			server.startElection()
		case <-time.After(server.heartbeatTimeout * time.Millisecond):
			server.appendEntry("")
		}
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
