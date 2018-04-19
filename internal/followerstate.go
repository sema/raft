package internal

import (
	"time"
	"github/sema/go-raft"
)

type followerState struct {
	storage go_raft.PersistentStorage
	volatileStorage VolatileStorage
	gateway go_raft.ServerGateway

	leaderElectionTimeout time.Duration
}

func NewFollowerState() ServerState {
	return &followerState{
		storage:               nil,
		gateway:               nil,
		leaderElectionTimeout: 0,
	}
}

func (f *followerState) Name() (name string) {
	return "follower"
}

func (f *followerState) Enter() () {

}

func (f *followerState) HandleRequestVote(request go_raft.RequestVoteRequest) (response go_raft.RequestVoteResponse, newState ServerState) {
	currentTerm := f.storage.CurrentTerm()

	f.tryVoteForCandidate(request)

	return go_raft.RequestVoteResponse{
		Term:        currentTerm,
		VoteGranted: f.storage.VotedFor() == request.CandidateID,
	}, nil
}

func (f *followerState) tryVoteForCandidate(request go_raft.RequestVoteRequest) {
	if !f.isCandidateLogReplicationUpToDate(request) {
		return
	}

	f.storage.SetVotedForIfUnset(request.CandidateID)
}

func (f *followerState) isCandidateLogReplicationUpToDate(request go_raft.RequestVoteRequest) bool {
	logEntry, ok := f.storage.LatestLogEntry()

	if !ok {
		return true // we have no logs locally, thus candidate can't be behind
	}

	if logEntry.Term < request.LastLogTerm {
		return true // candidate is at a newer term
	}

	if logEntry.Term == request.LastLogTerm && logEntry.Index <= request.LastLogIndex {
		return true // candidate has the same or more log entries for the current term
	}

	return false // candidate is at older term or has fewer entries
}

func (f *followerState) HandleAppendEntries(request go_raft.AppendEntriesRequest) (response go_raft.AppendEntriesResponse, nextState ServerState) {
	if !f.isLogConsistent(request) {
		return go_raft.AppendEntriesResponse{
			Term:    f.storage.CurrentTerm(),
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
		Term:    f.storage.CurrentTerm(),
		Success: true,
	}, nil
}

func (f *followerState) isLogConsistent(request go_raft.AppendEntriesRequest) bool {
	if request.PrevLogIndex == 0 && request.PrevLogTerm == 0 {
		// Base case - no previous log entries in log
		return true
	}

	logEntry, exists := f.storage.Log(request.PrevLogIndex)
	if exists && logEntry.Term == request.PrevLogTerm {
		// Induction case - previous log entry consistent
		return true
	}

	// Leader log is not consistent with local log
	return false
}

func (f *followerState) HandleLeaderElectionTimeout() (newState ServerState) {
	return NewCandidateState()
}

func (f *followerState) Exit() {

}