package actor

import (
	"log"
)

type candidateMode struct {
	persistentStorage Storage
	volatileStorage   *VolatileStorage
	config            Config

	ticksSinceLastHeartbeat  Tick
	ticksUntilLeaderElection Tick

	votes map[ServerID]bool
}

func newCandidateMode(persistentStorage Storage, volatileStorage *VolatileStorage, config Config) actorModeStrategy {
	return &candidateMode{
		persistentStorage: persistentStorage,
		volatileStorage:   volatileStorage,
		config:            config,
	}
}

func (s *candidateMode) PreExecuteModeChange(message Message) (newMode ActorMode, newTerm Term) {
	if message.Kind == msgAppendEntries {
		return FollowerMode, message.Term
	}

	return ExistingMode, 0
}

func (s *candidateMode) Process(message Message) (result *MessageResult) {
	switch message.Kind {
	case msgVoteFor:
		return s.handleRequestVote(message)
	case msgVoteForResponse:
		return s.handleRequestVoteResponse(message)
	case msgTick:
		return s.handleTick(message)
	}

	// Ignore message - allows us to add new messages in the future in a backwards compatible manner
	log.Printf("Unexpected message (%s) observed in Candidate mode", message.Kind)
	return newMessageResult()
}

func (s *candidateMode) Name() (name string) {
	return "CandidateMode"
}

func (s *candidateMode) handleTick(message Message) *MessageResult {
	s.ticksSinceLastHeartbeat++

	if s.ticksSinceLastHeartbeat >= s.ticksUntilLeaderElection {
		return s.startLeaderElection()
	}

	return newMessageResult()
}

func (s *candidateMode) startLeaderElection() *MessageResult {
	result := newMessageResult()
	result.ChangeMode(CandidateMode, s.persistentStorage.CurrentTerm()+1)
	return result
}

func (s *candidateMode) handleRequestVote(message Message) (result *MessageResult) {
	// Multiple candidates in same Term. stateContext always votes for itself when entering CandidateMode state, so
	// skip voting process.

	result = newMessageResult()
	result.WithMessage(NewMessageVoteForResponse(message.From, s.volatileStorage.ServerID, s.persistentStorage.CurrentTerm(), false))
	return result
}

func (s *candidateMode) handleRequestVoteResponse(message Message) (result *MessageResult) {
	s.votes[message.From] = message.VoteGranted

	votesPositive := 0
	votesNegative := 0
	for serverID := range s.votes {
		if s.votes[serverID] {
			votesPositive++
		} else {
			votesNegative++
		}
	}

	// We could fail fast if we detect a quorum of rejects. Only looking at the positive case
	// is enough for Raft to work however.
	if votesPositive >= s.config.Quorum() {
		result = newMessageResult()
		result.ChangeMode(LeaderMode, s.persistentStorage.CurrentTerm())
		return result
	}

	return newMessageResult()
}

func (s *candidateMode) Enter() (messagesOut []Message) {
	s.ticksSinceLastHeartbeat = 0
	s.ticksUntilLeaderElection = getTicksWithSplay(s.config.LeaderElectionTimeout, s.config.LeaderElectionTimeoutSplay)

	s.votes = make(map[ServerID]bool)

	// vote for ourselves
	s.votes[s.volatileStorage.ServerID] = true
	s.persistentStorage.SetVotedFor(s.volatileStorage.ServerID)

	if s.persistentStorage.VotedFor() != s.volatileStorage.ServerID {
		log.Panicf("Candidate could not vote for itself as it entered candidate mode, which it must do. Vote already cast for %s", s.persistentStorage.VotedFor())
	}

	// Send RPCs
	for _, serverID := range s.config.Servers {
		if serverID == s.volatileStorage.ServerID {
			continue // skip self
		}

		logEntry := s.persistentStorage.LatestLogEntry()

		messagesOut = append(messagesOut, NewMessageVoteFor(
			serverID,
			s.volatileStorage.ServerID,
			s.persistentStorage.CurrentTerm(),
			logEntry.Index,
			logEntry.Term))
	}

	return messagesOut
}

func (s *candidateMode) Exit() {
}
