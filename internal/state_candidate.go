package internal

import (
	"database/sql/driver"
	"github/sema/go-raft"
)

type candidateState struct {
	persistentStorage go_raft.PersistentStorage
	volatileStorage   *VolatileStorage
	gateway           go_raft.ServerGateway
	discovery         go_raft.Discovery
}

func NewCandidateState(persistentStorage go_raft.PersistentStorage, volatileStorage *VolatileStorage, gateway go_raft.ServerGateway, discovery go_raft.Discovery) ServerState {
	return &commonState{
		wrapped: &candidateState{
			persistentStorage: persistentStorage,
			volatileStorage:   volatileStorage,
			gateway:           gateway,
			discovery:         discovery,
		},
		persistentStorage: persistentStorage,
		volatileStorage:   volatileStorage,
		gateway:           gateway,
		discovery:         discovery,
	}
}

func (s *candidateState) Name() (name string) {
	return "candidate"
}

func (s *candidateState) HandleRequestVote(request go_raft.RequestVoteRequest) (response go_raft.RequestVoteResponse, newState ServerState) {
	// Multiple candidates in same term. StateContext always votes for itself when entering candidate state, so
	// skip voting process.

	return go_raft.RequestVoteResponse{
		Term:        s.persistentStorage.CurrentTerm(),
		VoteGranted: s.persistentStorage.VotedFor() == request.CandidateID,
	}, nil
}

func (s *candidateState) HandleAppendEntries(request go_raft.AppendEntriesRequest) (response go_raft.AppendEntriesResponse, nextState ServerState) {
	return go_raft.AppendEntriesResponse{}, NewFollowerState(s.persistentStorage, s.volatileStorage, s.gateway)
}

func (s *candidateState) TriggerLeaderElection() (newState ServerState) {
	panic("implement me")
}

// TODO send request vote RPCs to all other leaders
// If majority of hosts send votes then transition to leader

// TODO, remember, this may be triggered while already a candidate, should trigger new election
