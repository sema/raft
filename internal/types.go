package internal

import "github.com/sema/raft"

type StateImpl interface {
	Enter() (nextState raft.ServerState)
	RequestVote(raft.RequestVoteRequest) raft.RequestVoteResponse
	AppendEntries(raft.AppendEntriesRequest) raft.AppendEntriesResponse
	Exit()
}
