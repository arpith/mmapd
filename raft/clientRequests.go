package raft

import (
	"fmt"
)

type returnChanMessage struct {
	err  error
	json string
}

type readRequest struct {
	key        string
	returnChan chan returnChanMessage
}

type writeRequest struct {
	key        string
	value      string
	returnChan chan returnChanMessage
}

func (s *server) handleReadRequest(req readRequest) {
	fmt.Println("Got read request")
}

func (s *server) handleWriteRequest(req writeRequest) {
	fmt.Println("Got write request")
}
