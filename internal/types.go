package internal

import "github/sema/go-raft"

type ServerState interface {
	Name() string

	HandleRequestVote(go_raft.RequestVoteRequest) (response go_raft.RequestVoteResponse, newState ServerState)
	HandleAppendEntries(go_raft.AppendEntriesRequest) (response go_raft.AppendEntriesResponse, newState ServerState)
	TriggerLeaderElection() (newState ServerState)
}

type StateContext interface {
	RequestVote(go_raft.RequestVoteRequest) go_raft.RequestVoteResponse
	AppendEntries(go_raft.AppendEntriesRequest) go_raft.AppendEntriesResponse
	TriggerLeaderElection()
}
