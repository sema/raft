package actor_test

import (
	"testing"

	"github.com/sema/raft/pkg/actor"
	"github.com/stretchr/testify/assert"
)

func TestEnter__CandidateIncrementsTermAndSendsVoteForMessagesOnEnter(t *testing.T) {
	act, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	assert.Equal(t, actor.Term(0), storage.CurrentTerm())

	// Assertions for voteFor messages in test method
	testTransitionFromFollowerToCandidate(t, act, storage)

	assert.Equal(t, actor.Term(1), storage.CurrentTerm())
	assert.Equal(t, actor.CandidateMode, act.Mode())
}

func TestTick__CandidateRetriesCandidacyIfNoLeaderIsElected(t *testing.T) {
	act, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	testTransitionFromFollowerToCandidate(t, act, storage)

	assert.Equal(t, actor.Term(1), storage.CurrentTerm())

	testProgressTime(act, 11)

	assert.Equal(t, actor.Term(2), storage.CurrentTerm())
}

func TestAppendEntries__CandidateTransitionsToFollowerIfLeaderIsDetected(t *testing.T) {
	act, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	testTransitionFromFollowerToCandidate(t, act, storage)

	assert.Equal(t, actor.Term(1), storage.CurrentTerm())

	act.Process(actor.NewMessageAppendEntries(
		localServerID, peerServer1ID, actor.Term(1), 0, 0, 0, []actor.LogEntry{}))

	assert.Equal(t, actor.Term(1), storage.CurrentTerm())
	assert.Equal(t, actor.FollowerMode, act.Mode())
}

func TestTick__CandidateTransitionsToLeaderIfEnoughVotesSucceed(t *testing.T) {
	act, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	testTransitionFromFollowerToCandidate(t, act, storage)

	assert.Equal(t, actor.Term(1), storage.CurrentTerm())

	act.Process(actor.NewMessageVoteForResponse(
		localServerID, peerServer1ID, actor.Term(1), true))

	assert.Equal(t, actor.Term(1), storage.CurrentTerm())
	assert.Equal(t, actor.LeaderMode, act.Mode())
}

func TestTick__CandidateStaysACandidateIfVoteFails(t *testing.T) {
	act, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	testTransitionFromFollowerToCandidate(t, act, storage)

	assert.Equal(t, actor.Term(1), storage.CurrentTerm())

	act.Process(actor.NewMessageVoteForResponse(
		localServerID, peerServer1ID, actor.Term(1), false))
	act.Process(actor.NewMessageVoteForResponse(
		localServerID, peerServer2ID, actor.Term(1), false))

	assert.Equal(t, actor.Term(1), storage.CurrentTerm())
	assert.Equal(t, actor.CandidateMode, act.Mode())
}

func TestMsgVoteFor__IsIgnoredByCandidate(t *testing.T) {
	act, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	testTransitionFromFollowerToCandidate(t, act, storage)

	messagesExpected := []actor.Message{
		actor.NewMessageVoteForResponse(peerServer2ID, localServerID, actor.Term(1), false)}

	messagesOut := act.Process(
		actor.NewMessageVoteFor(localServerID, peerServer2ID, actor.Term(1), 0, 0))
	assert.Equal(t, messagesExpected, messagesOut)
	assert.Equal(t, storage.VotedFor(), localServerID)
}
