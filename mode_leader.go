package go_raft

import (
	"fmt"
	"sort"
)

type leaderMode struct {
	persistentStorage PersistentStorage
	volatileStorage   *VolatileStorage
	gateway           ServerGateway
	discovery         ServerDiscovery

	numTicksSinceLastHeartbeat int

	nextIndex  map[ServerID]LogIndex
	matchIndex map[ServerID]LogIndex
}

func newLeaderMode(persistentStorage PersistentStorage, volatileStorage *VolatileStorage, gateway ServerGateway, discovery ServerDiscovery) actorModeStrategy {
	return &leaderMode{
		persistentStorage:          persistentStorage,
		volatileStorage:            volatileStorage,
		gateway:                    gateway,
		discovery:                  discovery,
		numTicksSinceLastHeartbeat: 0,
	}
}

func (s *leaderMode) Name() (name string) {
	return "LeaderMode"
}

func (s *leaderMode) Enter() {
	s.numTicksSinceLastHeartbeat = 0
	s.broadcastHeartbeat()

	s.nextIndex = map[ServerID]LogIndex{}
	s.matchIndex = map[ServerID]LogIndex{}

	for _, serverID := range s.discovery.Servers() {
		s.nextIndex[serverID] = LogIndex(s.persistentStorage.LogLength() + 1)
		s.matchIndex[serverID] = 0
	}
}

func (s *leaderMode) PreExecuteModeChange(message Message) (newMode ActorMode, newTerm Term) {
	return ExistingMode, 0
}

func (s *leaderMode) Process(message Message) *MessageResult {
	switch message.Kind {
	case msgVoteFor:
		return s.handleRequestVote(message)
	case msgTick:
		return s.handleTick(message)
	case msgAppendEntriesResponse:
		return s.handleAppendEntriesResponse(message)
	default:
		panic(fmt.Sprintf("Unexpected Message %s passed to LeaderMode", message.Kind))
	}
}

func (s *leaderMode) handleTick(message Message) *MessageResult {
	s.numTicksSinceLastHeartbeat += 1

	if s.numTicksSinceLastHeartbeat > 4 {
		s.numTicksSinceLastHeartbeat = 0
		s.broadcastHeartbeat()
	}

	return newMessageResult()
}

func (s *leaderMode) handleAppendEntriesResponse(message Message) *MessageResult {
	if !message.Success {
		if s.matchIndex[message.From] >= LogIndex(0) {
			// TODO verify this is correct
			// We have already received a successful response and adjusted match/next index accordingly.
			// Skip this message.
			return newMessageResult()
		}

		s.nextIndex[message.From] = MaxLogIndex(s.nextIndex[message.From]-1, 1)
		s.heartbeat(message.From)
		return newMessageResult()
	}

	if s.matchIndex[message.From] > message.MatchIndex {
		s.matchIndex[message.From] = message.MatchIndex
		s.nextIndex[message.From] = message.MatchIndex + 1
	}

	return newMessageResult()
}

func (s *leaderMode) handleRequestVote(message Message) *MessageResult {
	// We are currently the LeaderMode of this Term, and incoming request is of the
	// same Term (otherwise we would have either changed state or rejected request).
	//
	// Lets just refrain from voting and hope the caller will turn into a FollowerMode
	// when we send the next heartbeat.
	return newMessageResult()
}

func (s *leaderMode) Exit() {

}

func (s *leaderMode) broadcastHeartbeat() {
	for _, serverID := range s.discovery.Servers() {
		if serverID == s.volatileStorage.ServerID {
			continue // skip self
		}
		s.heartbeat(serverID)
	}
}

func (s *leaderMode) heartbeat(targetServer ServerID) {

	// TODO not 100% sure about this one, should we take into account the size of entries?
	commitIndex := MinLogIndex(s.volatileStorage.CommitIndex, s.matchIndex[targetServer])

	currentIndex := s.nextIndex[targetServer] - 1
	logEntry, ok := s.persistentStorage.Log(currentIndex)
	if !ok {
		panic(fmt.Sprintf("Trying to lookup non-ExistingMode log entry (index: %d) during heartbeat", currentIndex))
	}

	s.gateway.Send(targetServer, newMessageAppendEntries(
		targetServer,
		s.volatileStorage.ServerID,
		s.persistentStorage.CurrentTerm(),

		commitIndex,

		logEntry.Index,
		logEntry.Term,

		// Entries  // TODO
	))
}

// TODO Implement protocol for deciding remote server position on initial election
// - success/failure feedback (bonus for actual position)

// TODO Maintain commit index
// - requires feedback on successful replication

// TODO API for adding new entries (essentially triggers new heartbeat)
// - NA

func (s *leaderMode) advanceCommitIndex() {

	matchIndexes := []int{}

	for _, index := range s.matchIndex {
		matchIndexes = append(matchIndexes, int(index))
	}

	sort.Ints(matchIndexes)
	quorum := len(s.discovery.Servers()) / 2

	quorumIndex := LogIndex(matchIndexes[quorum])

	// TODO not 100% sure about this part
	if s.volatileStorage.CommitIndex < quorumIndex {
		s.volatileStorage.CommitIndex = quorumIndex
	}
}
