package raft

import (
	"errors"
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
	s.db.ReadChan <- m
}

func (s *server) handleWriteRequest(req writeRequest) {
	key := req.key
	value := req.value
	command := "SET"
	c := make(chan bool)
	fmt.Println("GOT WRITE REQUEST!!!!")
	go s.appendEntry(command, key, value, c)
	isCommitted := <-c
	close(c)
	if isCommitted {
		m := db.WriteChanMessage{key, value, req.returnChan}
		s.db.WriteChan <- m
	} else {
		m := &db.ReturnChanMessage{errors.New("Couldn't Commit"), ""}
		req.returnChan <- *m
	}
}
