package raft

import (
	"fmt"
	"github.com/arpith/mmapd/db"
)

type returnChanMessage struct {
	err  error
	json string
}

type readRequest struct {
	key        string
	returnChan chan db.ReturnChanMessage
}

type writeRequest struct {
	key        string
	value      string
	returnChan chan db.ReturnChanMessage
}

func (s *server) handleReadRequest(req readRequest) {
	key := req.key
	c := req.returnChan
	m := db.ReadChanMessage{key, c}
	fmt.Println("Sending read request to db")
	s.db.ReadChan <- m
}

func (s *server) handleWriteRequest(req writeRequest) {
	/*
		Once a leader has been elected, it begins servicing
		client requests. Each client request contains a command to
		be executed by the replicated state machines. The leader
		appends the command to its log as a new entry, then issues
		AppendEntries RPCs in parallel to each of the other
		servers to replicate the entry. When the entry has been
		safely replicated (as described below), the leader applies
		the entry to its state machine and returns the result of that
		execution to the client. If followers crash or run slowly,
		or if network packets are lost, the leader retries AppendEntries
		RPCs indefinitely (even after it has responded to
		the client) until all followers eventually store all log entries
	*/
	key := req.key
	value := req.value
	command := "SET " + key + " " + value
	c := make(chan bool)
	fmt.Println("Going to append entry")
	s.appendEntry(command, c)
	fmt.Println("Waiting for response from server")
	<-c
	m := db.WriteChanMessage{key, value, req.returnChan}
	fmt.Println("Sending write request to db")
	s.db.WriteChan <- m
}
