package go_raft

import "fmt"

// LEADER
// TODO implement using advanced ticker?
// - upon election, send initial heartbeat
// - periodically send heartbeats
// - send append entries RPCs if behind - backtrack on error
// - update commit index (can be driven by append postprocess)

type leaderState struct {
	persistentStorage PersistentStorage
	volatileStorage   *VolatileStorage
	gateway           ServerGateway
	discovery         ServerDiscovery

	numTicksSinceLastHeartbeat int
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
	s.gateway.Send(targetServer, newCommandAppendEntries(
		targetServer,
		s.volatileStorage.ServerID,
		s.persistentStorage.CurrentTerm(),
		s.volatileStorage.CommitIndex, // TODO this is not true, needs to be addjusted according to target state
		LogIndex(0),                   // TODO
		Term(0),                       // TODO
		// Entries  // TODO
	))
}
