package go_raft

import (
	"time"
	"github/sema/go-raft/internal"
)

type Server interface {
	RequestVote(RequestVoteRequest) RequestVoteResponse
	AppendEntries(AppendEntriesRequest) AppendEntriesResponse
}

type serverImpl struct {
	state internal.ServerState

	storage Storage
	gateway ServerGateway

	mutex                 chan bool
	leaderElectionTimeout time.Duration

	ticker *time.Timer
}

func NewServer(storage Storage, gateway ServerGateway, nodeName NodeName, leaderElectionTimeout time.Duration) Server {
	mutex := make(chan bool, 1)
	mutex <- true // add token

	return &serverImpl{
		state: internal.NewFollowerState(),
		storage:          storage,
		gateway:          gateway,
		mutex:            mutex,
		leaderElectionTimeout: leaderElectionTimeout,
		ticker:                time.NewTimer(leaderElectionTimeout), // TODO decide on initial value
	}
}

func (si *serverImpl) Run() {
	for {
		select {
		case <-si.ticker.C:
			si.TriggerLeaderElectionTimeout()
		}
	}
}

func (si *serverImpl) RequestVote(request RequestVoteRequest) RequestVoteResponse {
	si.takeLock()
	defer si.releaseLock()

	currentTerm := si.storage.CurrentTerm()

	if request.CandidateTerm > currentTerm {
		// New candidate is initiating a vote - convert to follower and participate
		si.storage.SetCurrentTerm(request.CandidateTerm)
		currentTerm = request.CandidateTerm

		si.transitionState(internal.NewFollowerState())
	}

	if request.CandidateTerm < currentTerm {
		// Candidate belongs to previous term - ignore
		return RequestVoteResponse{
			Term:        currentTerm,
			VoteGranted: false,
		}
	}

	for {
		response, newState := si.state.HandleRequestVote(request)

		if newState != nil {
			si.transitionState(newState)
		} else {
			return response
		}
	}
}

func (si *serverImpl) AppendEntries(request AppendEntriesRequest) AppendEntriesResponse {
	si.takeLock()
	defer si.releaseLock()

	currentTerm := si.storage.CurrentTerm()

	if request.LeaderTerm > currentTerm {
		// New leader detected - follow it
		si.storage.SetCurrentTerm(request.LeaderTerm)
		currentTerm = request.LeaderTerm

		si.transitionState(internal.NewFollowerState())
	}

	if request.LeaderTerm < currentTerm {
		// Request from old leader - ignore request
		return AppendEntriesResponse{
			Term:    currentTerm,
			Success: false,
		}
	}

	si.leaderElectionTimeoutTimer.Reset(si.leaderElectionTimeout)

	for {
		response, newState := si.state.HandleAppendEntries(request)

		if newState != nil {
			si.transitionState(newState)
		} else {
			return response
		}
	}
}

func (si *serverImpl) TriggerLeaderElectionTimeout() {
	si.takeLock()
	defer si.releaseLock()

	for {
		newState := si.state.HandleLeaderElectionTimeout()

		if newState != nil {
			si.transitionState(newState)
		} else {
			return
		}
	}
}

func (si *serverImpl) takeLock() {
	<-si.mutex
}

func (si *serverImpl) releaseLock() {
	si.mutex <- true
}

func (si *serverImpl) transitionState(newState internal.ServerState) {
	si.state.Exit()
	si.state = newState
	si.state.Enter()
}

// LEADER
// TODO implement using advanced ticker?
// - upon election, send initial heartbeat
// - periodically send heartbeats
// - send append entries RPCs if behind - backtrack on error
// - update commit index (can be driven by append postprocess)
