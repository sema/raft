package go_raft

type candidateState struct {
	persistentStorage PersistentStorage
	volatileStorage   *VolatileStorage
	gateway           ServerGateway
	discovery         Discovery
}

func NewCandidateState(persistentStorage PersistentStorage, volatileStorage *VolatileStorage, gateway ServerGateway, discovery Discovery) serverState {
	return &commonState{
		wrapped: &candidateState{
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

func (s *candidateState) Name() (name string) {
	return "candidate"
}

func (s *candidateState) HandleRequestVote(request RequestVoteRequest) (response RequestVoteResponse, newState serverState) {
	// Multiple candidates in same term. stateContext always votes for itself when entering candidate state, so
	// skip voting process.

	return RequestVoteResponse{
		Term:        s.persistentStorage.CurrentTerm(),
		VoteGranted: s.persistentStorage.VotedFor() == request.CandidateID,
	}, nil
}

func (s *candidateState) HandleAppendEntries(request AppendEntriesRequest) (response AppendEntriesResponse, nextState serverState) {
	return AppendEntriesResponse{}, newFollowerState(s.persistentStorage, s.volatileStorage, s.gateway, s.discovery)
}

func (s *candidateState) TriggerLeaderElection() (newState serverState) {
	panic("implement me")
}

// TODO send request vote RPCs to all other leaders
// If majority of hosts send votes then transition to leader

// TODO, remember, this may be triggered while already a candidate, should trigger new election
