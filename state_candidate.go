package go_raft

type candidateState struct {
	persistentStorage PersistentStorage
	volatileStorage   *VolatileStorage
	gateway           ServerGateway
	discovery         Discovery
	context stateContext
}

func newCandidateState(persistentStorage PersistentStorage, volatileStorage *VolatileStorage, gateway ServerGateway, discovery Discovery, context stateContext) serverState {
	return &commonState{
		wrapped: &candidateState{
			persistentStorage: persistentStorage,
			volatileStorage:   volatileStorage,
			gateway:           gateway,
			discovery:         discovery,
			context: context,
		},
		persistentStorage: persistentStorage,
		volatileStorage:   volatileStorage,
		gateway:           gateway,
		discovery:         discovery,
		context: context,
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
	return AppendEntriesResponse{}, newFollowerState(s.persistentStorage, s.volatileStorage, s.gateway, s.discovery, s.context)
}

func (s *candidateState) TriggerLeaderElection() (newState serverState) {
	panic("implement me")
}

func (s *candidateState) Enter() {
	// TODO this needs to be changed - we can't have this other thread access data in the storage
	go s.requestVotes()
}

func (s *candidateState) Exit() {
	// TODO we need to to tear down the request votes thing
}

func (s *candidateState) requestVotes() {
	votesPositive := 0
	votesNegative := 0

	for _, serverID := range s.discovery.Servers() {
		if serverID == s.volatileStorage.ServerID {
			votesPositive += 1
			continue  // skip self
		}

		logEntry := s.persistentStorage.LatestLogEntry()

		response := s.gateway.SendRequestVoteRPC(serverID, RequestVoteRequest{
			CandidateTerm: s.persistentStorage.CurrentTerm(),
			CandidateID:   s.volatileStorage.ServerID,
			LastLogIndex:  logEntry.Index,
			LastLogTerm:   logEntry.Term,
		})

		if response.VoteGranted {
			votesPositive += 1
		} else {
			votesNegative += 1
		}
	}

	// TODO we should allow a leader to be elected using only a subset of votes
	// TODO add retries?
	if votesPositive > votesNegative {
		// Trigger state change
		s.context.TransitionStateIf(
			NewLeaderState(s.persistentStorage, s.volatileStorage, s.gateway, s.discovery, s.context),
			s.persistentStorage.CurrentTerm())
	}
}

// TODO send request vote RPCs to all other leaders
// If majority of hosts send votes then transition to leader

// TODO, remember, this may be triggered while already a candidate, should trigger new election
