package raft

import (
	"fmt"
	"github.com/arpith/mmapd/db"
	"io/ioutil"
	"strings"
	"time"
)

type server struct {
	id                  string
	state               string
	term                int
	votedFor            string
	db                  db.DB
	electionTimeout     timeout
	heartbeatTimeout    timeout
	config              []string
	commitIndex         int
	lastApplied         int
	nextIndex           []int
	matchIndex          []int
	voteRequests        chan voteRequest
	appendEntryRequests chan appendEntryRequest
}

func (s *server) listener() {
	for {
		select {
		case v := <-s.voteRequests:
			s.handleRequestForVote(v)
		case e := <-s.appendEntryRequests:
			s.handleAppendEntryRequest(e)
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

func Init(id string, db *db.DB) *server {
	configFilename := "config.txt"
	config := readConfig(configFilename)
	server := &server{
		id:                  id,
		state:               "follower",
		term:                0,
		votedFor:            "",
		db:                  *db,
		electionTimeout:     *createRandomTimeout(150, 300, time.Millisecond),
		heartbeatTimeout:    *createRandomTimeout(150, 300, time.Millisecond),
		config:              config,
		commitIndex:         0,
		lastApplied:         0,
		nextIndex:           make([]int, len(config)),
		matchIndex:          make([]int, len(config)),
		voteRequests:        make(chan voteRequest),
		appendEntryRequests: make(chan appendEntryRequest),
	}
	go server.listener()
	return server
}
