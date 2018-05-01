package raft_test

import (
	"github.com/golang/mock/gomock"
	"github.com/sema/go-raft"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEnter__CandidateIncrementsTermAndSendsVoteForMessagesOnEnter(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	assert.Equal(t, raft.Term(0), storage.CurrentTerm())

	// Assertions for voteFor messages in test method
	testTransitionFromFollowerToCandidate(actor, gatewayMock, storage)

	assert.Equal(t, raft.Term(1), storage.CurrentTerm())
	assert.Equal(t, raft.CandidateMode, actor.Mode())
}

func TestTick__CandidateRetriesCandidacyIfNoLeaderIsElected(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	testTransitionFromFollowerToCandidate(actor, gatewayMock, storage)

	assert.Equal(t, raft.Term(1), storage.CurrentTerm())

	gatewayMock.EXPECT().Send(gomock.Any(), gomock.Any()).Times(2)
	testProgressTime(actor, 11)

	assert.Equal(t, raft.Term(2), storage.CurrentTerm())
}

func TestAppendEntries__CandidateTransitionsToFollowerIfLeaderIsDetected(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	testTransitionFromFollowerToCandidate(actor, gatewayMock, storage)

	assert.Equal(t, raft.Term(1), storage.CurrentTerm())

	gatewayMock.EXPECT().Send(gomock.Any(), gomock.Any()).AnyTimes()

	actor.Process(raft.NewMessageAppendEntries(
		localServerID, peerServer1ID, raft.Term(1), 0, 0, 0, []raft.LogEntry{}))

	assert.Equal(t, raft.Term(1), storage.CurrentTerm())
	assert.Equal(t, raft.FollowerMode, actor.Mode())
}

func TestTick__CandidateTransitionsToLeaderIfEnoughVotesSucceed(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	testTransitionFromFollowerToCandidate(actor, gatewayMock, storage)

	assert.Equal(t, raft.Term(1), storage.CurrentTerm())

	gatewayMock.EXPECT().Send(gomock.Any(), gomock.Any()).AnyTimes()

	actor.Process(raft.NewMessageVoteForResponse(
		localServerID, peerServer1ID, raft.Term(1), true))

	assert.Equal(t, raft.Term(1), storage.CurrentTerm())
	assert.Equal(t, raft.LeaderMode, actor.Mode())
}

func TestTick__CandidateStaysACandidateIfVoteFails(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	testTransitionFromFollowerToCandidate(actor, gatewayMock, storage)

	assert.Equal(t, raft.Term(1), storage.CurrentTerm())

	actor.Process(raft.NewMessageVoteForResponse(
		localServerID, peerServer1ID, raft.Term(1), false))
	actor.Process(raft.NewMessageVoteForResponse(
		localServerID, peerServer2ID, raft.Term(1), false))

	assert.Equal(t, raft.Term(1), storage.CurrentTerm())
	assert.Equal(t, raft.CandidateMode, actor.Mode())
}

func TestMsgVoteFor__IsIgnoredByCandidate(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer2ID, raft.NewMessageVoteForResponse(
		peerServer2ID, localServerID, raft.Term(1), false))

	testTransitionFromFollowerToCandidate(actor, gatewayMock, storage)

	actor.Process(
		raft.NewMessageVoteFor(localServerID, peerServer2ID, raft.Term(1), 0, 0))

	assert.Equal(t, storage.VotedFor(), localServerID)
}
