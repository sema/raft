package internal

import "github/sema/go-raft"

type commonState struct {
	wrapped ServerState

	persistentStorage go_raft.PersistentStorage
	volatileStorage   *VolatileStorage
	gateway           go_raft.ServerGateway
	discovery         go_raft.Discovery
}

func (s *commonState) Name() string {
	return s.wrapped.Name()
}

func (*commonState) Enter() {
	panic("implement me")
}

func (s *commonState) HandleRequestVote(request go_raft.RequestVoteRequest) (response go_raft.RequestVoteResponse, newState ServerState) {
	currentTerm := s.persistentStorage.CurrentTerm()

	if request.CandidateTerm > currentTerm {
		// New candidate is initiating a vote - convert to follower and participate (Dissertation 3.3)
		s.persistentStorage.SetCurrentTerm(request.CandidateTerm)
		return go_raft.RequestVoteResponse{}, NewFollowerState(s.persistentStorage, s.volatileStorage, s.gateway, s.discovery)
	}

	if request.CandidateTerm < currentTerm {
		// Candidate belongs to previous term - ignore
		return go_raft.RequestVoteResponse{
			Term:        currentTerm,
			VoteGranted: false,
		}, nil
	}

	return s.wrapped.HandleRequestVote(request)
}

func (s *commonState) HandleAppendEntries(request go_raft.AppendEntriesRequest) (response go_raft.AppendEntriesResponse, newState ServerState) {
	currentTerm := s.persistentStorage.CurrentTerm()

	if request.LeaderTerm > currentTerm {
		// New leader detected - follow it (Dissertation 3.3)
		s.persistentStorage.SetCurrentTerm(request.LeaderTerm)
		return go_raft.AppendEntriesResponse{}, NewFollowerState()
	}

	if request.LeaderTerm < currentTerm {
		// Request from old leader - ignore request
		return go_raft.AppendEntriesResponse{
			Term:    currentTerm,
			Success: false,
		}, nil
	}

	return s.wrapped.HandleAppendEntries(request)
}

func (s *commonState) TriggerLeaderElection() ServerState {
	newTerm := s.persistentStorage.CurrentTerm() + 1
	s.persistentStorage.SetCurrentTerm(newTerm)
	s.persistentStorage.SetVotedForIfUnset(s.volatileStorage.ServerID)

	return NewCandidateState(s.persistentStorage, s.volatileStorage, s.gateway)
}

func (*commonState) Exit() {
	panic("implement me")
}
