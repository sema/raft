package go_raft

import "testing"

const candidate1ID = ServerID("candidate1.candidates.local")
const candidate2ID = ServerID("candidate2.candidates.local")

func TestRequestVoteIsAbleToGetVoteForInitialTerm(t *testing.T) {
	persistentStorage := NewMemoryStorage()
	persistentStorage.SetCurrentTerm(0)

	volatileStorage := &VolatileStorage{}
	discovery := NewStaticDiscovery([]ServerID{candidate1ID, candidate2ID})
	gateway := NewServerGatewayStub()

	state := newFollowerMode(persistentStorage, volatileStorage, gateway, discovery)

	response, newState := state.HandleRequestVote(RequestVoteRequest{
		CandidateTerm: 0,
		CandidateID:   candidate1ID,
		LastLogIndex:  0,
		LastLogTerm:   0,
	})

	equals(t, newState, nil)
	equals(t, response.VoteGranted, true)
	equals(t, persistentStorage.VotedFor(), candidate1ID)
}

func TestRequestVoteIsAbleToGetVoteForNonInitialTerm(t *testing.T) {
	persistentStorage := NewMemoryStorage()
	persistentStorage.SetCurrentTerm(2)

	volatileStorage := &VolatileStorage{}
	discovery := NewStaticDiscovery([]ServerID{candidate1ID})
	gateway := NewServerGatewayStub()

	persistentStorage.AppendLog(LogEntry{
		Term:  0,
		Index: 1,
	})
	persistentStorage.AppendLog(LogEntry{
		Term:  0,
		Index: 2,
	})
	persistentStorage.AppendLog(LogEntry{
		Term:  1,
		Index: 3,
	})

	state := newFollowerMode(persistentStorage, volatileStorage, gateway, discovery)

	response, newState := state.HandleRequestVote(RequestVoteRequest{
		CandidateTerm: 2,
		CandidateID:   candidate1ID,
		LastLogIndex:  3,
		LastLogTerm:   1,
	})

	equals(t, newState, nil)
	equals(t, response.VoteGranted, true)
	equals(t, persistentStorage.VotedFor(), candidate1ID)
}

func TestRequestVoteOnlyOneCandidateCanGetAVoteWithinATerm(t *testing.T) {

	persistentStorage := NewMemoryStorage()
	volatileStorage := &VolatileStorage{}
	discovery := NewStaticDiscovery([]ServerID{candidate1ID, candidate2ID})
	gateway := NewServerGatewayStub()

	persistentStorage.SetVotedForIfUnset(candidate1ID)

	state := newFollowerMode(persistentStorage, volatileStorage, gateway, discovery)

	response, newState := state.HandleRequestVote(RequestVoteRequest{
		CandidateTerm: 0,
		CandidateID:   candidate2ID,
		LastLogIndex:  0,
		LastLogTerm:   0,
	})

	equals(t, newState, nil)
	equals(t, response.VoteGranted, false)
	equals(t, persistentStorage.VotedFor(), candidate1ID)
}
