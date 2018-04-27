package go_raft

import (
	"fmt"
)

type followerMode struct {
	persistentStorage       PersistentStorage
	volatileStorage         *VolatileStorage
	gateway                 ServerGateway
	discovery               ServerDiscovery
	ticksSinceLastHeartbeat int
}

// TODO reuse the same state objects to reduce GC churn
func NewFollowerMode(persistentStorage PersistentStorage, volatileStorage *VolatileStorage, gateway ServerGateway, discovery ServerDiscovery) actorModeStrategy {
	return &followerMode{
		persistentStorage:       persistentStorage,
		volatileStorage:         volatileStorage,
		gateway:                 gateway,
		discovery:               discovery,
		ticksSinceLastHeartbeat: 0,
	}
}

func (s *followerMode) Name() (name string) {
	return "FollowerMode"
}

func (s *followerMode) Enter() {
	s.ticksSinceLastHeartbeat = 0
}

func (s *followerMode) PreExecuteModeChange(message Message) (newMode ActorMode, newTerm Term) {
	return ExistingMode, 0
}

func (s *followerMode) Process(message Message) *MessageResult {
	switch message.Kind {
	case msgAppendEntries:
		return s.handleAppendEntries(message)
	case msgVoteFor:
		return s.handleRequestVote(message)
	// TODO handle cmdVoteResponse
	case msgTick:
		return s.handleTick(message)
	default:
		panic(fmt.Sprintf("Unexpected Message %s passed to FollowerMode", message.Kind))
	}
}

func (s *followerMode) handleTick(message Message) *MessageResult {
	s.ticksSinceLastHeartbeat += 1

	// TODO change this into config
	// TODO add randomization
	if s.ticksSinceLastHeartbeat > 10 {
		return s.startLeaderElection()
	}

	return newMessageResult()
}

func (s *followerMode) startLeaderElection() *MessageResult {
	result := newMessageResult()
	result.ChangeMode(CandidateMode, s.persistentStorage.CurrentTerm()+1)
	return result
}

func (s *followerMode) handleRequestVote(message Message) *MessageResult {
	currentTerm := s.persistentStorage.CurrentTerm()

	s.tryVoteForCandidate(message.LastLogTerm, message.LastLogIndex, message.From)

	voteGranted := s.persistentStorage.VotedFor() == message.From

	s.gateway.Send(message.From, NewMessageVoteForResponse(message.From, s.volatileStorage.ServerID, currentTerm, voteGranted))
	return newMessageResult()
}

func (s *followerMode) handleAppendEntries(message Message) *MessageResult {
	s.ticksSinceLastHeartbeat = 0

	if !s.isLogConsistent(message.PreviousLogIndex, message.PreviousLogTerm) {
		s.gateway.Send(message.From, NewMessageAppendEntriesResponse(
			message.From,
			s.volatileStorage.ServerID,
			s.persistentStorage.CurrentTerm(),
			false,
			0))

		return newMessageResult()
	}

	// s.persistentStorage.MergeLogs(request.Entries)  // TODO

	if message.LeaderCommit > s.volatileStorage.CommitIndex {
		// It is possible that a newly elected LeaderMode has a lower commit index
		// than the previously elected LeaderMode. The commit index will eventually
		// reach the old point.
		//
		// In this implementation, we ensure the commit index never decreases
		// locally.
		s.volatileStorage.CommitIndex = message.LeaderCommit
	}

	s.persistentStorage.PruneLogEntriesAfter(message.PreviousLogIndex)
	s.persistentStorage.AppendLogs(message.LogEntries)

	logEntry := s.persistentStorage.LatestLogEntry()
	s.gateway.Send(message.From, NewMessageAppendEntriesResponse(
		message.From,
		s.volatileStorage.ServerID,
		s.persistentStorage.CurrentTerm(),
		true,
		logEntry.Index))

	return newMessageResult()
}

func (s *followerMode) tryVoteForCandidate(lastLogTerm Term, lastLogIndex LogIndex, candidateID ServerID) {
	if !s.isCandidateLogReplicationUpToDate(lastLogTerm, lastLogIndex) {
		return
	}

	s.persistentStorage.SetVotedForIfUnset(candidateID)
}

func (s *followerMode) isCandidateLogReplicationUpToDate(lastLogTerm Term, lastLogIndex LogIndex) bool {
	logEntry := s.persistentStorage.LatestLogEntry()

	if lastLogTerm < logEntry.Term {
		// Candidate is at an older Term
		return false
	}

	if lastLogTerm == logEntry.Term && lastLogIndex < logEntry.Index {
		// Candidate is at the same term, but has fewer log entries
		return false
	}

	// Candidate is up to date or newer
	return true
}

func (s *followerMode) isLogConsistent(prevLogIndex LogIndex, prevLogTerm Term) bool {
	if prevLogIndex == 0 && prevLogTerm == 0 {
		// Base case - no previous log entries in log
		return true
	}

	logEntry, exists := s.persistentStorage.Log(prevLogIndex)
	if exists && logEntry.Term == prevLogTerm {
		// Induction case - previous log entry consistent
		return true
	}

	// Leader log is not consistent with local log
	return false
}

func (s *followerMode) Exit() {

}
