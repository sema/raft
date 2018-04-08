package raft

import (
	"time"
)

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
	nodeName         NodeName

	// Volatile - leader
	nextIndex  LogIndex
	matchIndex LogIndex

	storage Storage
	gateway ServerGateway

	mutex                      chan bool
	leaderElectionTimeout      time.Duration
	leaderElectionTimeoutTimer *time.Timer
}

func NewServer(storage Storage, gateway ServerGateway, nodeName NodeName, leaderElectionTimeout time.Duration) Server {
	mutex := make(chan bool, 1)
	mutex <- true // add token

	return &serverImpl{
		nodeName:         nodeName,
		commitIndex:      0,
		lastAppliedIndex: 0,
		state:            Follower,
		nextIndex:        0,
		matchIndex:       0,
		storage:          storage,
		gateway:          gateway,
		mutex:            mutex,
		leaderElectionTimeout:      leaderElectionTimeout,
		leaderElectionTimeoutTimer: time.NewTimer(leaderElectionTimeout),
	}
}

func (si *serverImpl) Run() {
	for {
		select {
		case <-si.leaderElectionTimeoutTimer.C:
			si.handleLeaderElectionTimeout()
		}
	}
}

func (si *serverImpl) takeLock() {
	<-si.mutex
}

func (si *serverImpl) releaseLock() {
	si.mutex <- true
}

func (si *serverImpl) RequestVote(request RequestVoteRequest) RequestVoteResponse {
	si.takeLock()
	defer si.releaseLock()

	return si.handleRequestVote(request)
}

func (si *serverImpl) AppendEntries(request AppendEntriesRequest) AppendEntriesResponse {
	si.takeLock()
	defer si.releaseLock()

	return si.handleAppendEntries(request)
}

func (si *serverImpl) handleRequestVote(request RequestVoteRequest) RequestVoteResponse {
	currentTerm := si.storage.CurrentTerm()

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

func (si *serverImpl) handleAppendEntries(request AppendEntriesRequest) AppendEntriesResponse {
	currentTerm := si.storage.CurrentTerm()

	if request.leaderTerm > currentTerm {
		si.transitionToFollower(request.leaderTerm)
	}

	if si.state == Candidate {
		si.transitionToFollower(request.leaderTerm)
		// TODO remember to stop any in-flight votes etc
	}

	si.leaderElectionTimeoutTimer.Reset(si.leaderElectionTimeout)

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

func (si *serverImpl) handleLeaderElectionTimeout() {
	si.takeLock()
	defer si.releaseLock()

	newTerm := si.storage.CurrentTerm() + 1
	si.storage.SetCurrentTerm(newTerm)

	si.storage.SetVotedForIfUnset(si.nodeName)

	// TODO we should also ensure the channel is cleared
	si.leaderElectionTimeoutTimer.Reset(si.leaderElectionTimeout)

	si.state = Candidate

	// TODO send request vote RPCs to all other leaders
	// If majority of hosts send votes then transition to leader

	// TODO, remember, this may be triggered while already a candidate, should trigger new election
}

func (si *serverImpl) transitionToFollower(newTerm Term) {
	oldTerm := si.storage.CurrentTerm()
	si.storage.SetCurrentTerm(newTerm)

	if oldTerm != newTerm {
		// Don't clear vote if we are setting the same term multiple times, which might occur in certain edge cases
		si.storage.ClearVotedFor()
	}

	si.state = Follower
}

func (si *serverImpl) transitionToLeader() {

}

// LEADER
// TODO implement using advanced ticker?
// - upon election, send initial heartbeat
// - periodically send heartbeats
// - send append entries RPCs if behind - backtrack on error
// - update commit index (can be driven by append postprocess)
