package raft

import (
	"fmt"
	"github.com/arpith/mmapd/db"
)

func (s *server) commitEntries(leaderCommit int) {
	for i := s.commitIndex; i <= leaderCommit; i++ {
		if i > len(s.db.Log.Entries)-1 {
			fmt.Println("Can't commit an entry that's not in the log :)")
			return
		}
		if i == -1 {
			// s.commitIndex is initialised to -1
			// Will commit on the next iteration
			continue
		}
		entry := s.db.Log.Entries[i]
		c := make(chan db.ReturnChanMessage)
		m := db.WriteChanMessage{entry.Key, entry.Value, c}
		s.db.WriteChan <- m
		r := <-c
		if r.Err != nil {
			fmt.Println("Error committing entry: ", entry)
			fmt.Println(r.Err)
			break
		} else {
			s.commitIndex = i
		}
	}
}
