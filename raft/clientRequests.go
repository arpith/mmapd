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
	fmt.Println("Got write request")
}
