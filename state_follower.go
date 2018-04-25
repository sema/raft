package go_raft

import (
	"fmt"
)

type followerState struct {
	persistentStorage PersistentStorage
	volatileStorage   *VolatileStorage
	gateway           ServerGateway
	discovery         ServerDiscovery
	ticksSinceLastHeartbeat int
}

// TODO reuse the same state objects to reduce GC churn
func newFollowerState(persistentStorage PersistentStorage, volatileStorage *VolatileStorage, gateway ServerGateway, discovery ServerDiscovery) serverState {
	return &followerState{
			persistentStorage: persistentStorage,
			volatileStorage:   volatileStorage,
			gateway:           gateway,
			discovery:         discovery,
		ticksSinceLastHeartbeat: 0,
		}
}

func (s *followerState) Name() (name string) {
	return "follower"
}

func (s *followerState) Enter() {
	s.ticksSinceLastHeartbeat = 0
}

func (s *followerState) PreExecuteModeChange(command Command) (newMode interpreterMode, newTerm Term) {
	return existing, 0
}

func (s *followerState) Execute(command Command) *CommandResult {
	switch command.Kind {
	case cmdAppendEntries:
		return s.handleAppendEntries(command)
	case cmdVoteFor:
		return s.handleRequestVote(command)
	// TODO handle cmdVoteResponse
	case cmdTick:
		return s.handleTick(command)
	default:
		panic(fmt.Sprintf("Unexpected Command %s passed to follower", command.Kind))
	}
}

func (s *followerState) handleTick(command Command) *CommandResult {
	s.ticksSinceLastHeartbeat += 1

	// TODO change this into config
	// TODO add randomization
	if s.ticksSinceLastHeartbeat > 10 {
		return s.startLeaderElection()
	}

	return newCommandResult(true, s.persistentStorage.CurrentTerm())
}

func (s *followerState) startLeaderElection() *CommandResult {
	result := newCommandResult(true, s.persistentStorage.CurrentTerm())
	result.ChangeMode(candidate, s.persistentStorage.CurrentTerm() + 1)
	return result
}

func (s *followerState) handleRequestVote(command Command) *CommandResult {
	currentTerm := s.persistentStorage.CurrentTerm()

	s.tryVoteForCandidate(command.LastLogTerm, command.LastLogIndex, command.From)

	voteGranted := s.persistentStorage.VotedFor() == command.From

	s.gateway.SendRequestVoteResponseRPC(command.From, s.volatileStorage.ServerID, currentTerm, voteGranted)
	return newCommandResult(voteGranted, currentTerm)
}

func (s *followerState) handleAppendEntries(command Command) *CommandResult {
	if !s.isLogConsistent(command.PreviousLogIndex, command.PreviousLogTerm) {
		return newCommandResult(false, s.persistentStorage.CurrentTerm())
	}

	// s.persistentStorage.MergeLogs(request.Entries)  // TODO

	if command.LeaderCommit > s.volatileStorage.CommitIndex {
		// It is possible that a newly elected leader has a lower commit index
		// than the previously elected leader. The commit index will eventually
		// reach the old point.
		//
		// In this implementation, we ensure the commit index never decreases
		// locally.
		s.volatileStorage.CommitIndex = command.LeaderCommit
	}

	return newCommandResult(true, s.persistentStorage.CurrentTerm())
}

func (s *followerState) tryVoteForCandidate(lastLogTerm Term, lastLogIndex LogIndex, candidateID ServerID) {
	if !s.isCandidateLogReplicationUpToDate(lastLogTerm, lastLogIndex) {
		return
	}

	s.persistentStorage.SetVotedForIfUnset(candidateID)
}

func (s *followerState) isCandidateLogReplicationUpToDate(lastLogTerm Term, lastLogIndex LogIndex) bool {
	logEntry := s.persistentStorage.LatestLogEntry()

	if logEntry.Term < lastLogTerm {
		// Candidate is at a newer Term
		return true
	}

	if logEntry.Term == lastLogTerm && logEntry.Index <= lastLogIndex {
		// Candidate has the same or more log entries for the current Term.
		// Note: logEntry's Term and index are 0 if we do not have any logs locally yet
		return true
	}

	// Candidate is at older Term or has fewer entries
	return false
}

func (s *followerState) isLogConsistent(prevLogIndex LogIndex, prevLogTerm Term) bool {
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

func (s *followerState) Exit() {

}
