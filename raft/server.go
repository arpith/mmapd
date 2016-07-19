package main

import (
	"crypto/rand"
	"fmt"
	"github.com/arpith/mmapd/db"
	"io/ioutil"
	"math/big"
	"strings"
	"time"
)

type server struct {
	id                  string
	state               string
	term                int
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

func generateRandomInt(lower int, upper int) int {
	l := int64(lower)
	u := int64(upper)
	max := big.NewInt(u - l)
	r, err := rand.Int(rand.Reader, max)
	if err != nil {
		fmt.Println("Couldn't generate random int!")
	}
	return int(l + r.Int64())
}

func initServer(ip string, db *db.DB) *server {
	state := "follower"
	term := 0
	electionTimeoutPeriod := generateRandomInt(150, 300) * time.Millisecond
	heartbeatTimeoutPeriod := generateRandomInt(150, 300) * time.Millisecond
	electionTimeout := createTimeout(electionTimeoutPeriod)
	heartbeatTimeout := createTimeout(heartbeatTimeoutPeriod)
	config := readConfig("config.txt")
	voteChan := make(chan voteRequest)
	appendChan := make(chan appendEntryRequest)
	server := &server{ip, state, term, *db, *electionTimeout, *heartbeatTimeout, config, voteChan, appendChan}
	go server.listener()
	return server
}
