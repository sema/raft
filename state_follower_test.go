package go_raft

import "testing"

func TestRequestVoteIsAbleToGetVoteForInitialTerm(t *testing.T) {
	candidateID := ServerID("candidate1.candidates.local")

	persistentStorage := NewMemoryStorage()
	volatileStorage := &VolatileStorage{}
	discovery := NewStaticDiscovery([]ServerID{candidateID})
	gateway := NewServerGateway()

	state := newFollowerState(persistentStorage, volatileStorage, gateway, discovery)

	response, newState := state.HandleRequestVote(RequestVoteRequest{
		CandidateTerm: 0,
		CandidateID:   candidateID,
		LastLogIndex:  0,
		LastLogTerm:   0,
	})

	if newState != nil {
		t.Errorf("State changed unexpectedly to %s", newState.Name())
	}

	if !response.VoteGranted {
		t.Error("Vote not granted as expected")
	}

	if persistentStorage.VotedFor() != candidateID {
		t.Errorf("Unexpected candidate selected, expected: %s, was: %s", candidateID, persistentStorage.VotedFor())
	}
}

func TestRequestVoteIsAbleToGetVoteForNonInitialTerm(t *testing.T) {
	candidateID := ServerID("candidate1.candidates.local")

	persistentStorage := NewMemoryStorage()
	volatileStorage := &VolatileStorage{}
	discovery := NewStaticDiscovery([]ServerID{candidateID})
	gateway := NewServerGateway()

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

	state := newFollowerState(persistentStorage, volatileStorage, gateway, discovery)

	response, newState := state.HandleRequestVote(RequestVoteRequest{
		CandidateTerm: 2,
		CandidateID:   candidateID,
		LastLogIndex:  3,
		LastLogTerm:   1,
	})

	if newState != nil {
		t.Errorf("State changed unexpectedly to %s", newState.Name())
	}

	if !response.VoteGranted {
		t.Error("Vote not granted")
	}

	if persistentStorage.VotedFor() != candidateID {
		t.Errorf("Unexpected candidate selected, expected: %s, was: %s", candidateID, persistentStorage.VotedFor())
	}
}

func TestRequestVoteOnlyOneCandidateCanGetAVoteWithinATerm(t *testing.T) {
	candidate1ID := ServerID("candidate1.candidates.local")
	candidate2ID := ServerID("candidate2.candidates.local")

	persistentStorage := NewMemoryStorage()
	volatileStorage := &VolatileStorage{}
	discovery := NewStaticDiscovery([]ServerID{candidate1ID, candidate2ID})
	gateway := NewServerGateway()

	persistentStorage.SetVotedForIfUnset(candidate1ID)

	state := newFollowerState(persistentStorage, volatileStorage, gateway, discovery)

	response, newState := state.HandleRequestVote(RequestVoteRequest{
		CandidateTerm: 0,
		CandidateID:   candidate2ID,
		LastLogIndex:  0,
		LastLogTerm:   0,
	})

	if newState != nil {
		t.Errorf("State changed unexpectedly to %s", newState.Name())
	}

	if response.VoteGranted {
		t.Error("Vote was granted when other candidate should own the vote")
	}

	if persistentStorage.VotedFor() != candidate1ID {
		t.Errorf("Unexpected candidate selected, expected: %s, was: %s", candidate1ID, persistentStorage.VotedFor())
	}
}
