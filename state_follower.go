package go_raft

type followerState struct {
	persistentStorage PersistentStorage
	volatileStorage   *VolatileStorage
	gateway           ServerGateway
	discovery         Discovery
	context stateContext
}

// TODO reuse the same state objects to reduce GC churn
func newFollowerState(persistentStorage PersistentStorage, volatileStorage *VolatileStorage, gateway ServerGateway, discovery Discovery, context stateContext) serverState {
	return &commonState{
		wrapped: &followerState{
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

func (s *followerState) Name() (name string) {
	return "follower"
}

func (s *followerState) Enter() {

}

func (s *followerState) HandleRequestVote(request RequestVoteRequest) (response RequestVoteResponse, newState serverState) {
	currentTerm := s.persistentStorage.CurrentTerm()

	s.tryVoteForCandidate(request)

	return RequestVoteResponse{
		Term:        currentTerm,
		VoteGranted: s.persistentStorage.VotedFor() == request.CandidateID,
	}, nil
}

func (s *followerState) tryVoteForCandidate(request RequestVoteRequest) {
	if !s.isCandidateLogReplicationUpToDate(request) {
		return
	}

	s.persistentStorage.SetVotedForIfUnset(request.CandidateID)
}

func (s *followerState) isCandidateLogReplicationUpToDate(request RequestVoteRequest) bool {
	logEntry := s.persistentStorage.LatestLogEntry()

	if logEntry.Term < request.LastLogTerm {
		// Candidate is at a newer term
		return true
	}

	if logEntry.Term == request.LastLogTerm && logEntry.Index <= request.LastLogIndex {
		// Candidate has the same or more log entries for the current term.
		// Note: logEntry's term and index are 0 if we do not have any logs locally yet
		return true
	}

	// Candidate is at older term or has fewer entries
	return false
}

func (s *followerState) HandleAppendEntries(request AppendEntriesRequest) (response AppendEntriesResponse, nextState serverState) {
	if !s.isLogConsistent(request) {
		return AppendEntriesResponse{
			Term:    s.persistentStorage.CurrentTerm(),
			Success: false,
		}, nil
	}

	s.persistentStorage.MergeLogs(request.Entries)

	if request.LeaderCommit > s.volatileStorage.CommitIndex {
		// It is possible that a newly elected leader has a lower commit index
		// than the previously elected leader. The commit index will eventually
		// reach the old point.
		//
		// In this implementation, we ensure the commit index never decreases
		// locally.
		s.volatileStorage.CommitIndex = request.LeaderCommit
	}

	return AppendEntriesResponse{
		Term:    s.persistentStorage.CurrentTerm(),
		Success: true,
	}, nil
}

func (s *followerState) isLogConsistent(request AppendEntriesRequest) bool {
	if request.PrevLogIndex == 0 && request.PrevLogTerm == 0 {
		// Base case - no previous log entries in log
		return true
	}

	logEntry, exists := s.persistentStorage.Log(request.PrevLogIndex)
	if exists && logEntry.Term == request.PrevLogTerm {
		// Induction case - previous log entry consistent
		return true
	}

	// Leader log is not consistent with local log
	return false
}

func (s *followerState) TriggerLeaderElection() (newState serverState) {
	// Implemented in common context
	panic("implement me")
}

func (s *followerState) Exit() {

}
