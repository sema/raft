package go_raft

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
	discovery         Discovery
}

func NewLeaderState(persistentStorage PersistentStorage, volatileStorage *VolatileStorage, gateway ServerGateway, discovery Discovery) serverState {
	return &commonState{
		wrapped: &leaderState{
			persistentStorage: persistentStorage,
			volatileStorage:   volatileStorage,
			gateway:           gateway,
			discovery:         discovery,
		},
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

func (s *leaderState) HandleRequestVote(request RequestVoteRequest) (response RequestVoteResponse, newState serverState) {
	// We are currently the leader of this term, and incoming request is of the
	// same term (otherwise we would have either changed state or rejected request).
	//
	// Lets just refrain from voting and hope the caller will turn into a follower
	// when we send the next heartbeat.
	return RequestVoteResponse{
		Term:        s.persistentStorage.CurrentTerm(),
		VoteGranted: false,
	}, nil
}

func (s *leaderState) HandleAppendEntries(request AppendEntriesRequest) (response AppendEntriesResponse, nextState serverState) {
	// We should never be in this situation - we are the current leader of term X, while another
	// leader of the same term X is sending us heartbeats/appendEntry requests.
	//
	// There should never exist two leaders in the same term. For now, ignore other leader and maintain the split.
	// TODO handle protocol error
	return AppendEntriesResponse{
		Term:    s.persistentStorage.CurrentTerm(),
		Success: false,
	}, nil
}

func (s *leaderState) HandleLeaderElectionTimeout() (newState serverState) {
	// We are the current leader, thus this event is expected. No-op.
	return nil
}

func (s *leaderState) Exit() {

}

func (s *leaderState) heartbeat() {
	for serverID := range s.discovery.Servers() {
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
