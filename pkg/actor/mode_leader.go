package actor

import (
	"fmt"
	"log"
	"sort"
)

type leaderMode struct {
	persistentStorage Storage
	volatileStorage   *VolatileStorage
	config            Config

	numTicksSinceLastHeartbeat Tick

	nextIndex  map[ServerID]LogIndex
	matchIndex map[ServerID]LogIndex
	hasMatched map[ServerID]bool
}

func newLeaderMode(persistentStorage Storage, volatileStorage *VolatileStorage, config Config) actorModeStrategy {
	return &leaderMode{
		persistentStorage: persistentStorage,
		volatileStorage:   volatileStorage,
		config:            config,
	}
}

func (s *leaderMode) Name() (name string) {
	return "LeaderMode"
}

func (s *leaderMode) Enter() (messagesOut []Message) {
	s.numTicksSinceLastHeartbeat = 0

	s.nextIndex = map[ServerID]LogIndex{}
	s.matchIndex = map[ServerID]LogIndex{}
	s.hasMatched = map[ServerID]bool{}

	for _, serverID := range s.config.Servers {
		s.nextIndex[serverID] = LogIndex(s.persistentStorage.LatestLogEntry().Index + 1)
		s.matchIndex[serverID] = 0
		s.hasMatched[serverID] = false
	}

	s.matchIndex[s.volatileStorage.ServerID] = s.persistentStorage.LatestLogEntry().Index

	return s.createHeartbeatBroadcastMessages()
}

func (s *leaderMode) PreExecuteModeChange(message Message) (newMode ActorMode, newTerm Term) {
	return ExistingMode, 0
}

func (s *leaderMode) Process(message Message) *MessageResult {
	switch message.Kind {
	case msgTick:
		return s.handleTick(message)
	case msgAppendEntriesResponse:
		return s.handleAppendEntriesResponse(message)
	case msgProposal:
		return s.handleProposal(message)
	}

	// Ignore message - allows us to add new messages in the future in a backwards compatible manner
	log.Printf("Unexpected message (%s) observed in Leader mode", message.Kind)
	return newMessageResult()
}

func (s *leaderMode) handleTick(message Message) *MessageResult {
	s.numTicksSinceLastHeartbeat++

	if s.numTicksSinceLastHeartbeat < s.config.LeaderHeartbeatFrequency {
		return newMessageResult()
	}

	s.numTicksSinceLastHeartbeat = 0
	result := newMessageResult()
	result.WithMessages(s.createHeartbeatBroadcastMessages())

	return result
}

func (s *leaderMode) handleProposal(message Message) *MessageResult {
	s.persistentStorage.AppendLog(message.ProposalPayload)
	s.matchIndex[s.volatileStorage.ServerID]++

	return newMessageResult()
}

// Described in (3.5)
func (s *leaderMode) handleAppendEntriesResponse(message Message) *MessageResult {
	if !message.Success {
		if s.hasMatched[message.From] {
			// This is an out-of-order AppendEntries responses that we can safely discard. We have already recorded
			// a positive AppendEntries response and adjusted matchIndex/nextIndex accordingly. The current rejection
			// response must be a retry from before the acceptance response.
			return newMessageResult()
		}

		s.nextIndex[message.From] = MaxLogIndex(s.nextIndex[message.From]-1, 1)

		result := newMessageResult()
		result.WithMessage(s.createHeartbeatMessage(message.From))
		return result
	}

	s.hasMatched[message.From] = true

	if s.matchIndex[message.From] < message.MatchIndex {
		log.Printf("Setting matchIndex for %s to %d", message.From, message.MatchIndex)
		s.matchIndex[message.From] = message.MatchIndex
		s.nextIndex[message.From] = message.MatchIndex + 1
		s.advanceCommitIndex()
	}

	return newMessageResult()
}

func (s *leaderMode) Exit() {

}

func (s *leaderMode) createHeartbeatBroadcastMessages() []Message {
	var result []Message
	for _, serverID := range s.config.Servers {
		if serverID == s.volatileStorage.ServerID {
			continue // skip self
		}
		result = append(result, s.createHeartbeatMessage(serverID))
	}
	return result
}

func (s *leaderMode) createHeartbeatMessage(targetServer ServerID) Message {
	commitIndex := MinLogIndex(s.volatileStorage.CommitIndex, s.matchIndex[targetServer])

	currentIndex := s.nextIndex[targetServer] - 1
	logEntry, ok := s.persistentStorage.Log(currentIndex)
	if !ok {
		panic(fmt.Sprintf("Trying to lookup nonexisting log entry (index: %d) during heartbeat", currentIndex))
	}

	var logEntries []LogEntry
	if s.hasMatched[targetServer] {
		// BUG: The leader will send all missing log entries in a single message (unbounded) which will
		// cause failures in certain cases (adding new node, node unavailable for long periods). This issue will be
		// partially mitigated by snapshots/compaction. Adding a sensible max to number of log entries in a message
		// may also be useful.
		logEntries = s.persistentStorage.LogRange(logEntry.Index + 1)
	}

	return NewMessageAppendEntries(
		targetServer,
		s.volatileStorage.ServerID,
		s.persistentStorage.CurrentTerm(),

		commitIndex,

		logEntry.Index,
		logEntry.Term,

		logEntries,
	)
}

func (s *leaderMode) advanceCommitIndex() {
	var matchIndexes []int

	for _, index := range s.matchIndex {
		matchIndexes = append(matchIndexes, int(index))
	}

	sort.Ints(matchIndexes)

	quorum := s.config.Quorum()
	quorumIndex := LogIndex(matchIndexes[quorum-1])

	if s.volatileStorage.CommitIndex < quorumIndex {
		log.Printf("Increasing commitIndex to %d (%v)", quorumIndex, s.matchIndex)
		s.volatileStorage.CommitIndex = quorumIndex
	}
}
