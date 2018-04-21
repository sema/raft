package internal

import (
	"github/sema/go-raft"
)

type followerState struct {
	persistentStorage go_raft.PersistentStorage
	volatileStorage   *VolatileStorage
	gateway           go_raft.ServerGateway
	discovery         go_raft.Discovery
}

func NewFollowerState(persistentStorage go_raft.PersistentStorage, volatileStorage *VolatileStorage, gateway go_raft.ServerGateway, discovery go_raft.Discovery) ServerState {
	return &commonState{
		wrapped: &followerState{
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

func (s *followerState) Name() (name string) {
	return "follower"
}

func (s *followerState) Enter() {

}

func (s *followerState) HandleRequestVote(request go_raft.RequestVoteRequest) (response go_raft.RequestVoteResponse, newState ServerState) {
	currentTerm := s.persistentStorage.CurrentTerm()

	s.tryVoteForCandidate(request)

	return go_raft.RequestVoteResponse{
		Term:        currentTerm,
		VoteGranted: s.persistentStorage.VotedFor() == request.CandidateID,
	}, nil
}

func (s *followerState) tryVoteForCandidate(request go_raft.RequestVoteRequest) {
	if !s.isCandidateLogReplicationUpToDate(request) {
		return
	}

	s.persistentStorage.SetVotedForIfUnset(request.CandidateID)
}

func (s *followerState) isCandidateLogReplicationUpToDate(request go_raft.RequestVoteRequest) bool {
	logEntry, ok := s.persistentStorage.LatestLogEntry()

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

func (s *followerState) HandleAppendEntries(request go_raft.AppendEntriesRequest) (response go_raft.AppendEntriesResponse, nextState ServerState) {
	if !s.isLogConsistent(request) {
		return go_raft.AppendEntriesResponse{
			Term:    s.persistentStorage.CurrentTerm(),
			Success: false,
		}, nil
	}

	s.persistentStorage.MergeLogs(request.Entries)

	if request.LeaderCommit > s.volatileStorage.commitIndex {
		// It is possible that a newly elected leader has a lower commit index
		// than the previously elected leader. The commit index will eventually
		// reach the old point.
		//
		// In this implementation, we ensure the commit index never decreases
		// locally.
		s.volatileStorage.commitIndex = request.LeaderCommit
	}

	return go_raft.AppendEntriesResponse{
		Term:    s.persistentStorage.CurrentTerm(),
		Success: true,
	}, nil
}

func (s *followerState) isLogConsistent(request go_raft.AppendEntriesRequest) bool {
	if request.PrevLogIndex == 0 && request.PrevLogTerm == 0 {
		// Base case - no previous log entries in log
		return true
	}

	logEntry, exists := s.persistentStorage.Log(request.PrevLogIndex)
	if exists && logEntry.Term == request.PrevLogTerm {
		// Induction case - previous log entry consistent
		return true
	}

	// Leader log is not consistent with local log
	return false
}

func (s *followerState) TriggerLeaderElection() (newState ServerState) {
	// Implemented in common context
	panic("implement me")
}

func (s *followerState) Exit() {

}
