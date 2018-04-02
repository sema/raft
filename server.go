package raft

// AppendEntriesRequest contain the request payload for the AppendEntries RPC
type AppendEntriesRequest struct {
	leaderTerm   Term
	leaderID     NodeName
	leaderCommit LogIndex
	prevLogIndex LogIndex
	prevLogTerm  Term
	entries      []LogEntry
}

// AppendEntriesResponse contain the response payload for the AppendEntries RPC
type AppendEntriesResponse struct {
	success bool
	term    Term
}

// RequestVoteRequest contain the request payload for the RequestVote RPC
type RequestVoteRequest struct {
	candidateTerm Term
	candidateID   NodeName
	lastLogIndex  LogIndex
	lastLogTerm   Term
}

// RequestVoteResponse contain the response payload for the RequestVote RPC
type RequestVoteResponse struct {
	term        Term
	voteGranted bool
}

type Server interface {
	// Persisted data
	//currentTerm() Term
	//votedFor() NodeName
	// log[]

	RequestVote(RequestVoteRequest) RequestVoteResponse
	AppendEntries(AppendEntriesRequest) AppendEntriesResponse
}

type serverImpl struct {
	// Volatile
	commitIndex      LogIndex
	lastAppliedIndex LogIndex
	state            ServerState

	// Volatile - leader
	nextIndex  LogIndex
	matchIndex LogIndex

	storage Storage
}

func NewServer(storage Storage) Server {
	return &serverImpl{
		commitIndex:      0,
		lastAppliedIndex: 0,
		state:            Follower,
		nextIndex:        0,
		matchIndex:       0,
		storage:          storage,
	}
}

func (si *serverImpl) RequestVote(request RequestVoteRequest) RequestVoteResponse {
	currentTerm := si.storage.CurrentTerm()

	// TODO if request.term > currentTerm -> transition to follower and update currentTerm
	// TODO what to do about votedFor in this case? Should that be cleared?

	if request.candidateTerm > currentTerm {
		si.transitionToFollower(request.candidateTerm)
	}

	if request.candidateTerm < currentTerm {
		return RequestVoteResponse{
			term:        currentTerm,
			voteGranted: false,
		}
	}

	if !si.isCandidateLogUpToDate(request) {
		return RequestVoteResponse{
			term:        currentTerm,
			voteGranted: false,
		}
	}

	// TODO races beware
	si.storage.SetVotedForIfUnset(request.candidateID)

	return RequestVoteResponse{
		term:        currentTerm,
		voteGranted: si.storage.VotedFor() == request.candidateID,
	}
}

func (si *serverImpl) isCandidateLogUpToDate(request RequestVoteRequest) bool {
	logEntry, ok := si.storage.LatestLogEntry()

	if !ok {
		return true // we have no logs locally, thus candidate can't be behind
	}

	if logEntry.term < request.lastLogTerm {
		return true // candidate is at a newer term
	}

	if logEntry.term == request.lastLogTerm && logEntry.index <= request.lastLogIndex {
		return true // candidate has the same or more log entries for the current term
	}

	return false // candidate is at older term or has fewer entries
}

func (si *serverImpl) AppendEntries(request AppendEntriesRequest) AppendEntriesResponse {
	currentTerm := si.storage.CurrentTerm()

	if request.leaderTerm > currentTerm {
		si.transitionToFollower(request.leaderTerm)
	}

	if si.state == Candidate {
		si.transitionToFollower(request.leaderTerm)
	}

	// TODO reset election timeout

	if request.leaderTerm < currentTerm {
		// Leader is no longer the leader
		return AppendEntriesResponse{
			term:    currentTerm,
			success: false,
		}
	}

	if !si.isLogConsistent(request) {
		return AppendEntriesResponse{
			term:    currentTerm,
			success: false,
		}
	}

	si.storage.MergeLogs(request.entries)

	if request.leaderCommit > si.commitIndex {
		// It is possible that a newly elected leader has a lower commit index
		// than the previously elected leader. The commit index will eventually
		// reach the old point.
		//
		// In this implementation, we ensure the commit index never decreases
		// locally.
		si.commitIndex = request.leaderCommit
	}

	return AppendEntriesResponse{
		term:    currentTerm,
		success: true,
	}
}

func (si *serverImpl) isLogConsistent(request AppendEntriesRequest) bool {
	if request.prevLogIndex == 0 && request.prevLogTerm == 0 {
		// Base case - no previous log entries in log
		return true
	}

	logEntry, exists := si.storage.Log(request.prevLogIndex)
	if exists && logEntry.term == request.prevLogTerm {
		// Induction case - previous log entry consistent
		return true
	}

	// Leader log is not consistent with local log
	return false
}

func (si *serverImpl) transitionToFollower(newTerm Term) {
	si.storage.SetCurrentTerm(newTerm)
	si.state = Follower
}

func (l *lifecycle) run() {
	stateMap := map[ServerState]lifecycle2.State{
		Follower: lifecycle2.NewFollowerState(),
	}

	// TODO we could also have the function return a state, if we want to enforce the creation of new objects on each transition
	currentState := stateMap[Follower]

	for {
		nextState := currentState.Enter()
		currentState = stateMap[nextState]
	}

}
