package go_raft

import (
	"fmt"
)

type candidateMode struct {
	persistentStorage PersistentStorage
	volatileStorage   *VolatileStorage
	gateway           ServerGateway
	discovery         ServerDiscovery

	ticksSinceLastHeartbeat int

	votes map[ServerID]bool
}

func newCandidateMode(persistentStorage PersistentStorage, volatileStorage *VolatileStorage, gateway ServerGateway, discovery ServerDiscovery) actorModeStrategy {
	return &candidateMode{
		persistentStorage:       persistentStorage,
		volatileStorage:         volatileStorage,
		gateway:                 gateway,
		discovery:               discovery,
		ticksSinceLastHeartbeat: 0,
	}
}

func (s *candidateMode) PreExecuteModeChange(message Message) (newMode actorMode, newTerm Term) {
	if message.Kind == msgAppendEntries {
		return follower, message.Term
	}

	return existing, 0
}

func (s *candidateMode) Process(message Message) (result *MessageResult) {
	switch message.Kind {
	case msgVoteFor:
		return s.handleRequestVote(message)
	case msgVoteForResponse:
		return s.handleRequestVoteResponse(message)
	case msgTick:
		return s.handleTick(message)
	default:
		panic(fmt.Sprintf("Unexpected Message %s passed to candidate", message.Kind))
	}
}

func (s *candidateMode) Name() (name string) {
	return "candidate"
}

func (s *candidateMode) handleTick(message Message) *MessageResult {
	s.ticksSinceLastHeartbeat += 1

	if s.ticksSinceLastHeartbeat > 10 { // TODO move this into config
		return s.startLeaderElection()
	}

	return newMessageResult()
}

func (s *candidateMode) startLeaderElection() *MessageResult {
	result := newMessageResult()
	result.ChangeMode(candidate, s.persistentStorage.CurrentTerm()+1)
	return result
}

func (s *candidateMode) handleRequestVote(message Message) (result *MessageResult) {
	// Multiple candidates in same Term. stateContext always votes for itself when entering candidate state, so
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

	quorum := len(s.discovery.Servers()) / 2

	// TODO fail fast if majority rejects leader?
	if votesPositive > quorum {
		result = newMessageResult()
		result.ChangeMode(leader, s.persistentStorage.CurrentTerm())
		return result
	}

	return newMessageResult()
}

func (s *candidateMode) Enter() {
	s.ticksSinceLastHeartbeat = 0

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
	// TODO we need to to tear down the request votes thing
}

// TODO send request vote RPCs to all other leaders
// If majority of hosts send votes then transition to leader

// TODO, remember, this may be triggered while already a candidate, should trigger new election
