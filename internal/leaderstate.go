package internal

import (
	"github/sema/go-raft"
)

type leaderState struct {
	storage go_raft.Storage
	volatileStorage VolatileStorage
	gateway go_raft.ServerGateway
}

func NewLeaderState() ServerState {
	return &leaderState{
		storage:               nil,
		gateway:               nil,
	}
}

func (f *leaderState) Name() (name string) {
	return "leader"
}

func (f *leaderState) Enter() () {

}

func (f *leaderState) HandleRequestVote(request go_raft.RequestVoteRequest) (response go_raft.RequestVoteResponse, newState ServerState) {
	currentTerm := f.storage.CurrentTerm()

	if request.CandidateTerm > currentTerm {
		// New leader detection happening - Bump to new term and convert to follower
		f.storage.SetCurrentTerm(request.CandidateTerm)
		return go_raft.RequestVoteResponse{}, NewFollowerState()
	}

	if request.CandidateTerm < currentTerm {
		// Reject request as it belongs to previous term
		return go_raft.RequestVoteResponse{
			Term:        currentTerm,
			VoteGranted: false,
		}, nil
	}

	// We are currently the leader of this term, and incoming request is of the
	// same term (otherwise we would have either changed state or rejected request).
	//
	// Lets just refrain from voting and hope the caller will turn into a follower
	// when we send the next heartbeat.
	return go_raft.RequestVoteResponse{
		Term:        currentTerm,
		VoteGranted: false,
	}, nil
}

func (f *leaderState) HandleAppendEntries(request go_raft.AppendEntriesRequest) (response go_raft.AppendEntriesResponse, nextState ServerState) {
	currentTerm := f.storage.CurrentTerm()

	if request.LeaderTerm > currentTerm {
		// New leader elected - Bump to new term and convert to follower
		f.storage.SetCurrentTerm(request.LeaderTerm)
		return go_raft.AppendEntriesResponse{}, NewFollowerState()
	}

	if request.LeaderTerm < currentTerm {
		// Reject request as it belongs to previous term
		return go_raft.AppendEntriesResponse{
			Term:    currentTerm,
			Success: false,
		}, nil
	}

	// We should never be in this situation - we are the current leader of term X, while another
	// leader of the same term X is sending us heartbeats/appendEntry requests.
	//
	// There should never exist two leaders in the same term. For now, ignore other leader and maintain the split.
	// TODO handle protocol error
	return go_raft.AppendEntriesResponse{
		Term:    currentTerm,
		Success: false,
	}, nil
}

func (f *leaderState) HandleLeaderElectionTimeout() (newState ServerState) {
	// We are the current leader, thus this event is expected. No-op.
	return nil
}

func (f *leaderState) Exit() {

}
