package internal

import (
	"time"

	"github.com/sema/raft"
)

type followerState struct {
	storage raft.Storage
	gateway raft.ServerGateway

	leaderElectionTimeout time.Duration
}

func NewFollowerState() StateImpl {
	return &followerState{}
}

func (f *followerState) Enter() (newState raft.ServerState) {

}

func (f *followerState) RequestVote(raft.RequestVoteRequest) raft.RequestVoteResponse {

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

func (f *followerState) AppendEntries(raft.AppendEntriesRequest) raft.AppendEntriesResponse {

}

func (f *followerState) Exit() {

}
