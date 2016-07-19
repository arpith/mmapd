package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"syscall"
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
}

func (s *server) handleWriteRequest(req writeRequest) {
}
