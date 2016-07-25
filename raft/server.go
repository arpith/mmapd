package raft

import (
	"encoding/json"
	"fmt"
	"github.com/arpith/mmapd/db"
	"io/ioutil"
	"time"
)

type server struct {
	id               string
	state            string
	term             int
	votedFor         string
	db               db.DB
	electionTimeout  timeout
	heartbeatTimeout timeout
	config           []string
	commitIndex      int
	lastApplied      int
	nextIndex        []int
	matchIndex       []int
	voteRequests     chan voteRequest
	appendRequests   chan appendRequest
	writeRequests    chan writeRequest
	readRequests     chan readRequest
}

func (s *server) listener() {
	for {
		select {
		case v := <-s.voteRequests:
			fmt.Println("Got vote request")
			s.handleRequestForVote(v)
		case e := <-s.appendRequests:
			fmt.Println("Got append entry request")
			s.handleAppendEntryRequest(e)
		case r := <-s.readRequests:
			s.handleReadRequest(r)
		case w := <-s.writeRequests:
			s.handleWriteRequest(w)
		case <-s.heartbeatTimeout.ticker:
			if s.state == "leader" {
				fmt.Println("Going to send heartbeats")
				c := make(chan bool)
				go s.appendEntry("", c)
			}
		case <-s.electionTimeout.ticker:
			fmt.Println("Going to start election")
			go s.startElection()
		}
	}
}

func readConfig(filename string) []string {
	var config []string
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Couldn't read config file")
	}
	err = json.Unmarshal(content, &config)
	if err != nil {
		fmt.Println("Error unmarshalling config file: ", err)
	}
	return config
}

func Init(id string, configFilename string, db *db.DB) *server {
	config := readConfig(configFilename)
	server := &server{
		id:               id,
		state:            "follower",
		term:             0,
		votedFor:         "",
		db:               *db,
		electionTimeout:  *createRandomTimeout(150, 300, time.Millisecond),
		heartbeatTimeout: *createRandomTimeout(150, 300, time.Millisecond),
		config:           config,
		commitIndex:      0,
		lastApplied:      0,
		nextIndex:        make([]int, len(config)),
		matchIndex:       make([]int, len(config)),
		voteRequests:     make(chan voteRequest),
		appendRequests:   make(chan appendRequest),
		writeRequests:    make(chan writeRequest),
		readRequests:     make(chan readRequest),
	}
	go server.listener()
	return server
}
