package go_raft_test

import (
	"testing"
	"github.com/sema/go-raft"
	"github.com/stretchr/testify/assert"
	"github.com/golang/mock/gomock"
	"github.com/sema/go-raft/mocks"
)

func testTransitionFromFollowerToCandidate(actor go_raft.Actor, gatewayMock *mock_go_raft.MockServerGateway) {
	gatewayMock.EXPECT().Send(peerServer1ID, go_raft.NewMessageVoteFor(
		peerServer1ID, localServerID, go_raft.Term(1), 0, 0))
	gatewayMock.EXPECT().Send(peerServer2ID, go_raft.NewMessageVoteFor(
		peerServer2ID, localServerID, go_raft.Term(1), 0, 0))

	testProgressTime(actor, 11)
}

func testProgressTime(actor go_raft.Actor, numTicks int) {
	for i := 0; i < numTicks; i++ {
		actor.Process(go_raft.NewMessageTick(localServerID, localServerID))
	}
}

func TestEnter__CandidateIncrementsTermAndSendsVoteForMessagesOnEnter(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	assert.Equal(t, go_raft.Term(0), storage.CurrentTerm())

	// Assertions for voteFor messages in test method
	testTransitionFromFollowerToCandidate(actor, gatewayMock)

	assert.Equal(t, go_raft.Term(1), storage.CurrentTerm())
	assert.Equal(t, go_raft.CandidateMode, actor.Mode())
}

func TestTick__CandidateRetriesCandidacyIfNoLeaderIsElected(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	testTransitionFromFollowerToCandidate(actor, gatewayMock)

	assert.Equal(t, go_raft.Term(1), storage.CurrentTerm())

	gatewayMock.EXPECT().Send(gomock.Any(), gomock.Any()).Times(2)
	testProgressTime(actor, 11)

	assert.Equal(t, go_raft.Term(2), storage.CurrentTerm())
}

func TestAppendEntries__CandidateTransitionsToFollowerIfLeaderIsDetected(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	testTransitionFromFollowerToCandidate(actor, gatewayMock)

	assert.Equal(t, go_raft.Term(1), storage.CurrentTerm())

	gatewayMock.EXPECT().Send(gomock.Any(), gomock.Any()).AnyTimes()

	actor.Process(go_raft.NewMessageAppendEntries(
		localServerID, peerServer1ID, go_raft.Term(1), 0, 0, 0, []go_raft.LogEntry{}))

	assert.Equal(t, go_raft.Term(1), storage.CurrentTerm())
	assert.Equal(t, go_raft.FollowerMode, actor.Mode())
}

func TestTick__CandidateTransitionsToLeaderIfEnoughVotesSucceed(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	testTransitionFromFollowerToCandidate(actor, gatewayMock)

	assert.Equal(t, go_raft.Term(1), storage.CurrentTerm())

	gatewayMock.EXPECT().Send(gomock.Any(), gomock.Any()).AnyTimes()

	actor.Process(go_raft.NewMessageVoteForResponse(
		localServerID, peerServer1ID, go_raft.Term(1), true))

	assert.Equal(t, go_raft.Term(1), storage.CurrentTerm())
	assert.Equal(t, go_raft.LeaderMode, actor.Mode())
}

func TestTick__CandidateStaysACandidateIfVoteFails(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	testTransitionFromFollowerToCandidate(actor, gatewayMock)

	assert.Equal(t, go_raft.Term(1), storage.CurrentTerm())

	actor.Process(go_raft.NewMessageVoteForResponse(
		localServerID, peerServer1ID, go_raft.Term(1), false))
	actor.Process(go_raft.NewMessageVoteForResponse(
		localServerID, peerServer2ID, go_raft.Term(1), false))

	assert.Equal(t, go_raft.Term(1), storage.CurrentTerm())
	assert.Equal(t, go_raft.CandidateMode, actor.Mode())
}

func TestMsgVoteFor__IsIgnoredByCandidate(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer2ID, go_raft.NewMessageVoteForResponse(
		peerServer2ID, localServerID, go_raft.Term(1), false))

	testTransitionFromFollowerToCandidate(actor, gatewayMock)

	actor.Process(
		go_raft.NewMessageVoteFor(localServerID, peerServer2ID, go_raft.Term(1), 0, 0))

	assert.Equal(t, storage.VotedFor(), localServerID)
}
