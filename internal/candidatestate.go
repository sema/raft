package internal

import (
	"github/sema/go-raft"
)

type candidateState struct {
	storage go_raft.Storage
	volatileStorage VolatileStorage
	gateway go_raft.ServerGateway
}

func NewCandidateState() ServerState {
	return &candidateState{
		storage:               nil,
		gateway:               nil,
	}
}

func (f *candidateState) Name() (name string) {
	return "candidate"
}

func (f *candidateState) Enter() () {

}

func (f *candidateState) HandleRequestVote(request go_raft.RequestVoteRequest) (response go_raft.RequestVoteResponse, newState ServerState) {
	// Multiple candidates in same term. Server always votes for itself when entering candidate state, so
	// skip voting process.

	return go_raft.RequestVoteResponse{
		Term:        f.storage.CurrentTerm(),
		VoteGranted: f.storage.VotedFor() == request.CandidateID,
	}, nil
}

func (f *candidateState) HandleAppendEntries(request go_raft.AppendEntriesRequest) (response go_raft.AppendEntriesResponse, nextState ServerState) {

	// TODO determine response here? This indicates that a leader was elected in current term and we need to
	// convert to follower

}

func (f *candidateState) HandleLeaderElectionTimeout() (newState ServerState) {
	newTerm := f.storage.CurrentTerm() + 1
	f.storage.SetCurrentTerm(newTerm)

	f.storage.SetVotedForIfUnset(f.nodeName)

	// TODO send request vote RPCs to all other leaders
	// If majority of hosts send votes then transition to leader

	// TODO, remember, this may be triggered while already a candidate, should trigger new election
}

func (f *candidateState) Exit() {

}