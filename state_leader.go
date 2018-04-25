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
}

func newLeaderState(persistentStorage PersistentStorage, volatileStorage *VolatileStorage, gateway ServerGateway, discovery ServerDiscovery) serverState {
	return &leaderState{
			persistentStorage: persistentStorage,
			volatileStorage:   volatileStorage,
			gateway:           gateway,
			discovery:         discovery,
		}
}

func (s *leaderState) Name() (name string) {
	return "leader"
}

func (s *leaderState) Enter() {

}

func (s *leaderState) PreExecuteModeChange(command Command) (newMode interpreterMode, newTerm Term) {
	return existing, 0
}

func (s *leaderState) Execute(command Command) *CommandResult {
	switch command.Kind {
	case cmdVoteFor:
		return s.handleRequestVote(command)
	case cmdTick:
		return newCommandResult(true, s.persistentStorage.CurrentTerm())  // noop
	default:
		panic(fmt.Sprintf("Unexpected Command %s passed to candidate", command.Kind))
	}
}


func (s *leaderState) handleRequestVote(command Command) *CommandResult {
	// We are currently the leader of this Term, and incoming request is of the
	// same Term (otherwise we would have either changed state or rejected request).
	//
	// Lets just refrain from voting and hope the caller will turn into a follower
	// when we send the next heartbeat.
	return newCommandResult(false, s.persistentStorage.CurrentTerm())
}

func (s *leaderState) Exit() {

}

func (s *leaderState) heartbeat() {
	for _, serverID := range s.discovery.Servers() {
		go s.sendSingleHeartbeat(serverID)
	}

}

func (s *leaderState) sendSingleHeartbeat(targetServer ServerID) {
	s.gateway.SendAppendEntriesRPC(
		targetServer,
		AppendEntriesRequest{
			LeaderTerm:   s.persistentStorage.CurrentTerm(),
			LeaderID:     s.volatileStorage.ServerID,
			LeaderCommit: s.volatileStorage.CommitIndex, // TODO this is not true, needs to be addjusted according to target state
			PrevLogIndex: LogIndex(0),                   // TODO
			PrevLogTerm:  Term(0),                       // TODO
			// Entries  // TODO
		})

}

func (s *leaderState) TriggerLeaderElection() (newState serverState) {
	panic("implement me")
}
