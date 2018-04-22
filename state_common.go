package go_raft

type commonState struct {
	wrapped serverState

	persistentStorage PersistentStorage
	volatileStorage   *VolatileStorage
	gateway           ServerGateway
	discovery         Discovery
}

func (s *commonState) Name() string {
	return s.wrapped.Name()
}

func (*commonState) Enter() {
	panic("implement me")
}

func (s *commonState) HandleRequestVote(request RequestVoteRequest) (response RequestVoteResponse, newState serverState) {
	currentTerm := s.persistentStorage.CurrentTerm()

	if request.CandidateTerm > currentTerm {
		// New candidate is initiating a vote - convert to follower and participate (Dissertation 3.3)
		s.persistentStorage.SetCurrentTerm(request.CandidateTerm)
		return RequestVoteResponse{}, newFollowerState(s.persistentStorage, s.volatileStorage, s.gateway, s.discovery)
	}

	if request.CandidateTerm < currentTerm {
		// Candidate belongs to previous term - ignore
		return RequestVoteResponse{
			Term:        currentTerm,
			VoteGranted: false,
		}, nil
	}

	return s.wrapped.HandleRequestVote(request)
}

func (s *commonState) HandleAppendEntries(request AppendEntriesRequest) (response AppendEntriesResponse, newState serverState) {
	currentTerm := s.persistentStorage.CurrentTerm()

	if request.LeaderTerm > currentTerm {
		// New leader detected - follow it (Dissertation 3.3)
		s.persistentStorage.SetCurrentTerm(request.LeaderTerm)
		return AppendEntriesResponse{}, newFollowerState(s.persistentStorage, s.volatileStorage, s.gateway, s.discovery)
	}

	if request.LeaderTerm < currentTerm {
		// Request from old leader - ignore request
		return AppendEntriesResponse{
			Term:    currentTerm,
			Success: false,
		}, nil
	}

	return s.wrapped.HandleAppendEntries(request)
}

func (s *commonState) TriggerLeaderElection() serverState {
	newTerm := s.persistentStorage.CurrentTerm() + 1
	s.persistentStorage.SetCurrentTerm(newTerm)
	s.persistentStorage.SetVotedForIfUnset(s.volatileStorage.ServerID)

	return NewCandidateState(s.persistentStorage, s.volatileStorage, s.gateway, s.discovery)
}

func (*commonState) Exit() {
	panic("implement me")
}
