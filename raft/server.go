package main

import (
	"crypto/rand"
	"fmt"
	"github.com/arpith/mmapd/db"
)

type server struct {
	id                  string
	term                int
	db                  db.DB
	electionTimeout     int
	heartbeatTimeout    int
	config              []string
	commitIndex         int
	lastApplied         int
	nextIndex           []int
	matchIndex          []int
	voteRequests        chan voteRequest
	appendEntryRequests chan appendEntryRequest
}

func (server *server) listener() {
	for {
		select {
		case v := <-server.voteRequests:
			server.handleRequestForVote(v)
		case e := <-server.appendEntryRequests:
			server.handleAppendEntryRequest(e)
		case <-s.heartbeatTimeout.ticker:
			s.appendEntry("")
		case <-s.electionTimeout.ticker:
			s.startElection()
		}
	}
}

func readConfig(filename string) []string {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Couldn't read config file")
	}
	servers := strings.Split(string(content), "\n")
	return servers
}

func initServer(ip string, db *db.DB) *server {
	state := "follower"
	term := &term{0, false, 0}
	electionTimeout := 150 + rand.Int(rand.Reader, 150)
	heartbeatTimeout := 150 + rand.Int(rand.Reader, 150)
	config = readConfig("config.txt")
	voteChan := make(chan voteRequest)
	appendChan := make(chan appendEntryRequest)
	server := &server{ip, state, term, electionTimeout, heartbeatTimeout, config, voteChan, appendChan}
	go server.listener()
	return server
}
