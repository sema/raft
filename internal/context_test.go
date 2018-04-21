package internal

import (
	"testing"
)

func TestRequestVoteIsAbleToGetVoteForInitialTerm(t *testing.T) {
	candidateID := NodeName("candidate1.candidates.local")

	storage := NewMemoryStorage()
	server := NewContext(storage)

	response := server.RequestVote(RequestVoteRequest{
		candidateTerm: 0,
		candidateID:   candidateID,
		lastLogIndex:  0,
		lastLogTerm:   0,
	})

	if !response.voteGranted {
		t.Error("Vote not granted")
	}

	if storage.VotedFor() != candidateID {
		t.Errorf("Unexpected candidate selected, expected: %s, was: %s", candidateID, storage.VotedFor())
	}
}

func TestRequestVoteIsAbleToGetVoteForNonInitialTerm(t *testing.T) {
	candidateID := NodeName("candidate1.candidates.local")

	storage := NewMemoryStorage()
	storage.AppendLog(LogEntry{
		term:  0,
		index: 1,
	})
	storage.AppendLog(LogEntry{
		term:  0,
		index: 2,
	})
	storage.AppendLog(LogEntry{
		term:  1,
		index: 3,
	})

	server := NewContext(storage)

	response := server.RequestVote(RequestVoteRequest{
		candidateTerm: 2,
		candidateID:   candidateID,
		lastLogIndex:  3,
		lastLogTerm:   1,
	})

	if !response.voteGranted {
		t.Error("Vote not granted")
	}

	if storage.VotedFor() != candidateID {
		t.Errorf("Unexpected candidate selected, expected: %s, was: %s", candidateID, storage.VotedFor())
	}
}

func TestRequestVoteOnlyOneCandidateCanGetAVoteWithinATerm(t *testing.T) {
	candidate1ID := NodeName("candidate1.candidates.local")
	candidate2ID := NodeName("candidate2.candidates.local")

	storage := NewMemoryStorage()
	storage.SetVotedForIfUnset(candidate1ID)
	server := NewContext(storage)

	response := server.RequestVote(RequestVoteRequest{
		candidateTerm: 0,
		candidateID:   candidate2ID,
		lastLogIndex:  0,
		lastLogTerm:   0,
	})

	if response.voteGranted {
		t.Error("Vote was granted when other candidate should own the vote")
	}

	if storage.VotedFor() != candidate1ID {
		t.Errorf("Unexpected candidate selected, expected: %s, was: %s", candidate1ID, storage.VotedFor())
	}
}
