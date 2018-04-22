package go_raft

import "testing"

type mockStateContext struct {}

func (c mockStateContext) TransitionStateIf(newState serverState, term Term) {

}

func (c mockStateContext) RequestVote(request RequestVoteRequest) RequestVoteResponse {
	return RequestVoteResponse{}
}

func (c mockStateContext) AppendEntries(request AppendEntriesRequest) AppendEntriesResponse {
	return AppendEntriesResponse{}
}

func (c mockStateContext) TriggerLeaderElection() {}

func (c mockStateContext) CurrentStateName() string {
	return ""
}


func TestThreadsafeStateContext_proxiesCallsToWrappedContext(t *testing.T) {
	context := newThreadsafeStateContext(mockStateContext{})

	context.AppendEntries(AppendEntriesRequest{})
	context.RequestVote(RequestVoteRequest{})
	context.TriggerLeaderElection()
	context.CurrentStateName()
}
