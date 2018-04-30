package go_raft

import (
	"log"
)

type candidateMode struct {
	persistentStorage PersistentStorage
	volatileStorage   *VolatileStorage
	gateway           ServerGateway
	discovery         ServerDiscovery
	config Config

	ticksSinceLastHeartbeat Tick
	ticksUntilLeaderElection Tick

	votes map[ServerID]bool
}

func newCandidateMode(persistentStorage PersistentStorage, volatileStorage *VolatileStorage, gateway ServerGateway, discovery ServerDiscovery, config Config) actorModeStrategy {
	return &candidateMode{
		persistentStorage:       persistentStorage,
		volatileStorage:         volatileStorage,
		gateway:                 gateway,
		discovery:               discovery,
		config: config,
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
	s.ticksSinceLastHeartbeat += 1

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

	s.gateway.Send(message.From, NewMessageVoteForResponse(message.From, s.volatileStorage.ServerID, s.persistentStorage.CurrentTerm(), false))
	return newMessageResult()
}

func (s *candidateMode) handleRequestVoteResponse(message Message) (result *MessageResult) {
	s.votes[message.From] = message.VoteGranted

	votesPositive := 0
	votesNegative := 0
	for serverID := range s.votes {
		if s.votes[serverID] {
			votesPositive += 1
		} else {
			votesNegative += 1
		}
	}

	// We could fail fast if we detect a quorum of rejects. Only looking at the positive case
	// is enough for Raft to work however.
	if votesPositive >= s.discovery.Quorum() {
		result = newMessageResult()
		result.ChangeMode(LeaderMode, s.persistentStorage.CurrentTerm())
		return result
	}

	return newMessageResult()
}

func (s *candidateMode) Enter() {
	s.ticksSinceLastHeartbeat = 0
	s.ticksUntilLeaderElection = getTicksWithSplay(s.config.LeaderElectionTimeout, s.config.LeaderElectionTimeoutSplay)

	s.votes = make(map[ServerID]bool)

	// vote for ourselves
	s.votes[s.volatileStorage.ServerID] = true
	s.persistentStorage.SetVotedForIfUnset(s.volatileStorage.ServerID) // TODO ensure this happens

	// Send RPCs
	for _, serverID := range s.discovery.Servers() {
		if serverID == s.volatileStorage.ServerID {
			continue // skip self
		}

		logEntry := s.persistentStorage.LatestLogEntry()

		// TODO retries?
		s.gateway.Send(serverID, NewMessageVoteFor(serverID, s.volatileStorage.ServerID, s.persistentStorage.CurrentTerm(), logEntry.Index, logEntry.Term))
	}
}

func (s *candidateMode) Exit() {
}
