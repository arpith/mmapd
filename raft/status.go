package raft

type status struct {
	ID          string `json:"id"`
	State       string `json:"state"`
	Term        int    `json:"term"`
	VotedFor    string `json:"voted for"`
	CommitIndex int    `json:"commit index"`
	LastApplied int    `json:"last applied"`
	NextIndex   []int  `json:"next index"`
	MatchIndex  []int  `json:"match index"`
}
