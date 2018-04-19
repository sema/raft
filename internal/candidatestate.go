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
	currentTerm := f.storage.CurrentTerm()

	if request.CandidateTerm > currentTerm {
		f.storage.SetCurrentTerm(request.CandidateTerm)
		currentTerm = request.CandidateTerm
	}

	if request.CandidateTerm < currentTerm {
		return go_raft.RequestVoteResponse{
			Term:        currentTerm,
			VoteGranted: false,
		}, nil
	}

	f.tryVoteForCandidate(request)

	return go_raft.RequestVoteResponse{
		Term:        currentTerm,
		VoteGranted: f.storage.VotedFor() == request.CandidateID,
	}, nil
}

func (f *candidateState) HandleAppendEntries(request go_raft.AppendEntriesRequest) (response go_raft.AppendEntriesResponse, nextState ServerState) {
	currentTerm := f.storage.CurrentTerm()

	if request.LeaderTerm > currentTerm {
		f.storage.SetCurrentTerm(request.LeaderTerm)
		currentTerm = request.LeaderTerm
	}

	if request.LeaderTerm < currentTerm {
		// Leader is no longer the leader
		return go_raft.AppendEntriesResponse{
			Term:    currentTerm,
			Success: false,
		}, nil
	}

	if !f.isLogConsistent(request) {
		return go_raft.AppendEntriesResponse{
			Term:    currentTerm,
			Success: false,
		}, nil
	}

	f.storage.MergeLogs(request.Entries)

	if request.LeaderCommit > f.volatileStorage.commitIndex {
		// It is possible that a newly elected leader has a lower commit index
		// than the previously elected leader. The commit index will eventually
		// reach the old point.
		//
		// In this implementation, we ensure the commit index never decreases
		// locally.
		f.volatileStorage.commitIndex = request.LeaderCommit
	}

	return go_raft.AppendEntriesResponse{
		Term:    currentTerm,
		Success: true,
	}, nil
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