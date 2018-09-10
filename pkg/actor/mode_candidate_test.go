package actor_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sema/raft/pkg/actor"
	"github.com/stretchr/testify/assert"
)

func TestEnter__CandidateIncrementsTermAndSendsVoteForMessagesOnEnter(t *testing.T) {
	act, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	assert.Equal(t, actor.Term(0), storage.CurrentTerm())

	// Assertions for voteFor messages in test method
	testTransitionFromFollowerToCandidate(act, gatewayMock, storage)

	assert.Equal(t, actor.Term(1), storage.CurrentTerm())
	assert.Equal(t, actor.CandidateMode, act.Mode())
}

func TestTick__CandidateRetriesCandidacyIfNoLeaderIsElected(t *testing.T) {
	act, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	testTransitionFromFollowerToCandidate(act, gatewayMock, storage)

	assert.Equal(t, actor.Term(1), storage.CurrentTerm())

	gatewayMock.EXPECT().Send(gomock.Any(), gomock.Any()).Times(2)
	testProgressTime(act, 11)

	assert.Equal(t, actor.Term(2), storage.CurrentTerm())
}

func TestAppendEntries__CandidateTransitionsToFollowerIfLeaderIsDetected(t *testing.T) {
	act, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	testTransitionFromFollowerToCandidate(act, gatewayMock, storage)

	assert.Equal(t, actor.Term(1), storage.CurrentTerm())

	gatewayMock.EXPECT().Send(gomock.Any(), gomock.Any()).AnyTimes()

	act.Process(actor.NewMessageAppendEntries(
		localServerID, peerServer1ID, actor.Term(1), 0, 0, 0, []actor.LogEntry{}))

	assert.Equal(t, actor.Term(1), storage.CurrentTerm())
	assert.Equal(t, actor.FollowerMode, act.Mode())
}

func TestTick__CandidateTransitionsToLeaderIfEnoughVotesSucceed(t *testing.T) {
	act, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	testTransitionFromFollowerToCandidate(act, gatewayMock, storage)

	assert.Equal(t, actor.Term(1), storage.CurrentTerm())

	gatewayMock.EXPECT().Send(gomock.Any(), gomock.Any()).AnyTimes()

	act.Process(actor.NewMessageVoteForResponse(
		localServerID, peerServer1ID, actor.Term(1), true))

	assert.Equal(t, actor.Term(1), storage.CurrentTerm())
	assert.Equal(t, actor.LeaderMode, act.Mode())
}

func TestTick__CandidateStaysACandidateIfVoteFails(t *testing.T) {
	act, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	testTransitionFromFollowerToCandidate(act, gatewayMock, storage)

	assert.Equal(t, actor.Term(1), storage.CurrentTerm())

	act.Process(actor.NewMessageVoteForResponse(
		localServerID, peerServer1ID, actor.Term(1), false))
	act.Process(actor.NewMessageVoteForResponse(
		localServerID, peerServer2ID, actor.Term(1), false))

	assert.Equal(t, actor.Term(1), storage.CurrentTerm())
	assert.Equal(t, actor.CandidateMode, act.Mode())
}

func TestMsgVoteFor__IsIgnoredByCandidate(t *testing.T) {
	act, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer2ID, actor.NewMessageVoteForResponse(
		peerServer2ID, localServerID, actor.Term(1), false))

	testTransitionFromFollowerToCandidate(act, gatewayMock, storage)

	act.Process(
		actor.NewMessageVoteFor(localServerID, peerServer2ID, actor.Term(1), 0, 0))

	assert.Equal(t, storage.VotedFor(), localServerID)
}
