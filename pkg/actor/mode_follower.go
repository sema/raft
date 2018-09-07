package actor

import "log"

type followerMode struct {
	persistentStorage Storage
	volatileStorage   *VolatileStorage
	gateway           ServerGateway
	config            Config

	ticksSinceLastHeartbeat  Tick
	ticksUntilLeaderElection Tick
}

func NewFollowerMode(persistentStorage Storage, volatileStorage *VolatileStorage, gateway ServerGateway, config Config) actorModeStrategy {
	return &followerMode{
		persistentStorage: persistentStorage,
		volatileStorage:   volatileStorage,
		gateway:           gateway,
		config:            config,
	}
}

func (s *followerMode) Name() (name string) {
	return "FollowerMode"
}

func (s *followerMode) Enter() {
	s.ticksSinceLastHeartbeat = 0
	s.ticksUntilLeaderElection = getTicksWithSplay(s.config.LeaderElectionTimeout, s.config.LeaderElectionTimeoutSplay)
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
	case msgTick:
		return s.handleTick(message)
	}

	// Ignore message - allows us to add new messages in the future in a backwards compatible manner
	log.Printf("Unexpected message (%s) observed in Follower mode", message.Kind)
	return newMessageResult()
}

func (s *followerMode) handleTick(message Message) *MessageResult {
	s.ticksSinceLastHeartbeat++

	if s.ticksSinceLastHeartbeat >= s.ticksUntilLeaderElection {
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

	s.persistentStorage.MergeLogs(message.LogEntries)

	if message.LeaderCommit > s.volatileStorage.CommitIndex {
		// It is possible that a newly elected LeaderMode has a lower commit index
		// than the previously elected LeaderMode. The commit index will eventually
		// reach the old point.
		//
		// In this implementation, we ensure the commit index never decreases
		// locally.
		s.volatileStorage.CommitIndex = message.LeaderCommit
	}

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

	if s.persistentStorage.VotedFor() == NoVote {
		s.persistentStorage.SetVotedFor(candidateID)
	}
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
