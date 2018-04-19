package internal

import "github/sema/go-raft"

type ServerState interface {
	Name() string

	Enter()

	HandleRequestVote(go_raft.RequestVoteRequest) (response go_raft.RequestVoteResponse, newState ServerState)
	HandleAppendEntries(go_raft.AppendEntriesRequest) (response go_raft.AppendEntriesResponse, newState ServerState)
	HandleLeaderElectionTimeout() (newState ServerState)

	Exit()
}
