package go_raft

import (
	"fmt"
	"sort"
)

type leaderState struct {
	persistentStorage PersistentStorage
	volatileStorage   *VolatileStorage
	gateway           ServerGateway
	discovery         ServerDiscovery

	numTicksSinceLastHeartbeat int

	nextIndex  map[ServerID]LogIndex
	matchIndex map[ServerID]LogIndex
}

func newLeaderState(persistentStorage PersistentStorage, volatileStorage *VolatileStorage, gateway ServerGateway, discovery ServerDiscovery) serverState {
	return &leaderState{
			persistentStorage: persistentStorage,
			volatileStorage:   volatileStorage,
			gateway:           gateway,
			discovery:         discovery,
		    numTicksSinceLastHeartbeat: 0,
		}
}

func (s *leaderState) Name() (name string) {
	return "leader"
}

func (s *leaderState) Enter() {
	s.numTicksSinceLastHeartbeat = 0
	s.broadcastHeartbeat()

	s.nextIndex = map[ServerID]LogIndex{}
	s.matchIndex = map[ServerID]LogIndex{}

	for _, serverID := range s.discovery.Servers() {
		s.nextIndex[serverID] = LogIndex(s.persistentStorage.LogLength() + 1)
		s.matchIndex[serverID] = 0
	}
}

func (s *leaderState) PreExecuteModeChange(command Command) (newMode interpreterMode, newTerm Term) {
	return existing, 0
}

func (s *leaderState) Execute(command Command) *CommandResult {
	switch command.Kind {
	case cmdVoteFor:
		return s.handleRequestVote(command)
	case cmdTick:
		return s.handleTick(command)
	case cmdAppendEntriesResponse:
		return s.handleAppendEntriesResponse(command)
	default:
		panic(fmt.Sprintf("Unexpected Command %s passed to leader", command.Kind))
	}
}

func (s *leaderState) handleTick(command Command) *CommandResult {
	s.numTicksSinceLastHeartbeat += 1

	if s.numTicksSinceLastHeartbeat > 4 {
		s.numTicksSinceLastHeartbeat = 0
		s.broadcastHeartbeat()
	}

	return newCommandResult()
}

func (s *leaderState) handleAppendEntriesResponse(command Command) *CommandResult {
	if !command.Success {
		if s.matchIndex[command.From] >= LogIndex(0) {
			// TODO verify this is correct
			// We have already received a successful response and adjusted match/next index accordingly.
			// Skip this command.
			return newCommandResult()
		}

		s.nextIndex[command.From] = MaxLogIndex(s.nextIndex[command.From] - 1, 1)
		s.heartbeat(command.From)
		return newCommandResult()
	}

	if s.matchIndex[command.From] > command.MatchIndex {
		s.matchIndex[command.From] = command.MatchIndex
		s.nextIndex[command.From] = command.MatchIndex + 1
	}

	return newCommandResult()
}

func (s *leaderState) handleRequestVote(command Command) *CommandResult {
	// We are currently the leader of this Term, and incoming request is of the
	// same Term (otherwise we would have either changed state or rejected request).
	//
	// Lets just refrain from voting and hope the caller will turn into a follower
	// when we send the next heartbeat.
	return newCommandResult()
}

func (s *leaderState) Exit() {

}

func (s *leaderState) broadcastHeartbeat() {
	for _, serverID := range s.discovery.Servers() {
		if serverID == s.volatileStorage.ServerID {
			continue  // skip self
		}
		s.heartbeat(serverID)
	}
}

func (s *leaderState) heartbeat(targetServer ServerID) {

	// TODO not 100% sure about this one, should we take into account the size of entries?
	commitIndex := MinLogIndex(s.volatileStorage.CommitIndex, s.matchIndex[targetServer])

	currentIndex := s.nextIndex[targetServer] - 1
	logEntry, ok := s.persistentStorage.Log(currentIndex)
	if !ok {
		panic(fmt.Sprintf("Trying to lookup non-existing log entry (index: %d) during heartbeat", currentIndex))
	}

	s.gateway.Send(targetServer, newCommandAppendEntries(
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

func (s *leaderState) advanceCommitIndex() {

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