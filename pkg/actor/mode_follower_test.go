package actor_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sema/raft/pkg/actor"
	"github.com/stretchr/testify/assert"
)

func TestMsgVoteFor__IsAbleToGetAVote(t *testing.T) {
	act, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer1ID, actor.NewMessageVoteForResponse(
		peerServer1ID, localServerID, actor.Term(1), true))

	act.Process(
		actor.NewMessageVoteFor(localServerID, peerServer1ID, actor.Term(1), 0, 0))

	assert.Equal(t, storage.VotedFor(), peerServer1ID)
}

func TestMsgVoteFor__DoesNotOverwriteExistingVoteInTerm(t *testing.T) {
	act, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	storage.SetCurrentTerm(actor.Term(1))
	storage.SetVotedFor(peerServer1ID)

	gatewayMock.EXPECT().Send(peerServer2ID, actor.NewMessageVoteForResponse(
		peerServer2ID, localServerID, actor.Term(1), false))

	act.Process(
		actor.NewMessageVoteFor(localServerID, peerServer2ID, actor.Term(1), 0, 0))

	assert.Equal(t, storage.VotedFor(), peerServer1ID)
}

func TestMsgTick__FollowerEventuallyTransitionsToCandidateAfterXTicks(t *testing.T) {
	act, gatewayMock, _, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(gomock.Any(), gomock.Any()).AnyTimes()

	testProgressTime(act, 11)

	assert.Equal(t, actor.CandidateMode, act.Mode())
}

func TestMsgVoteFor__IsRejectedIfIfCandidateLogBelongsToOlderTerm(t *testing.T) {
	act, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer1ID, actor.NewMessageVoteForResponse(
		peerServer1ID, localServerID, actor.Term(1), false))

	storage.SetCurrentTerm(1)
	storage.AppendLog("") // term 1, index 1
	storage.AppendLog("") // term 1, index 2

	act.Process(
		actor.NewMessageVoteFor(localServerID, peerServer1ID, actor.Term(1), 0, 0))
}

func TestMsgVoteFor__IsRejectedIfIfCandidateLogIsBehindOnIndex(t *testing.T) {
	act, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer1ID, actor.NewMessageVoteForResponse(
		peerServer1ID, localServerID, actor.Term(1), false))

	storage.SetCurrentTerm(1)
	storage.AppendLog("") // term 1, index 1
	storage.AppendLog("") // term 1, index 2

	act.Process(
		actor.NewMessageVoteFor(localServerID, peerServer1ID, actor.Term(1), 1, 1))
}

func TestMsgAppendEntries__IsRejectedIfPreviousTermAndIndexDoesNotMatch(t *testing.T) {
	act, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer1ID, actor.NewMessageAppendEntriesResponse(
		peerServer1ID, localServerID, actor.Term(1), false, actor.LogIndex(0)))

	storage.SetCurrentTerm(1)
	storage.AppendLog("") // term 1, index 1
	storage.AppendLog("") // term 1, index 2

	act.Process(
		actor.NewMessageAppendEntries(
			localServerID, peerServer1ID, actor.Term(1), 0, actor.LogIndex(3), actor.Term(3), []actor.LogEntry{}))
}

func TestMsgAppendEntries__IsAcceptedIfPreviousTermAndIndexMatch(t *testing.T) {
	act, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer1ID, actor.NewMessageAppendEntriesResponse(
		peerServer1ID, localServerID, actor.Term(1), true, actor.LogIndex(2)))

	storage.SetCurrentTerm(1)
	storage.AppendLog("") // term 1, index 1
	storage.AppendLog("") // term 1, index 2

	act.Process(
		actor.NewMessageAppendEntries(
			localServerID, peerServer1ID, actor.Term(1), 0, actor.LogIndex(2), actor.Term(1), []actor.LogEntry{}))
}

func TestMsgAppendEntries__AppendsNewEntriesToLog(t *testing.T) {
	act, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer1ID, actor.NewMessageAppendEntriesResponse(
		peerServer1ID, localServerID, actor.Term(1), true, actor.LogIndex(4)))

	storage.SetCurrentTerm(1)
	storage.AppendLog("") // term 1, index 1
	storage.AppendLog("") // term 1, index 2

	act.Process(
		actor.NewMessageAppendEntries(
			localServerID, peerServer1ID, actor.Term(1), 0, actor.LogIndex(2), actor.Term(1),
			[]actor.LogEntry{
				actor.NewLogEntry(1, 3, ""),
				actor.NewLogEntry(1, 4, ""),
			}))

	assert.Equal(t, actor.LogIndex(4), storage.LatestLogEntry().Index)
}

func TestMsgAppendEntries__AppendingPreviouslyAppendedEntriesRetainsCurrentState(t *testing.T) {
	act, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer1ID, actor.NewMessageAppendEntriesResponse(
		peerServer1ID, localServerID, actor.Term(2), true, actor.LogIndex(4)))

	storage.SetCurrentTerm(1)
	storage.AppendLog("") // term 1, index 1
	storage.AppendLog("") // term 1, index 2

	storage.SetCurrentTerm(2)
	storage.AppendLog("") // term 2, index 3
	storage.AppendLog("") // term 2, index 4

	act.Process(
		actor.NewMessageAppendEntries(
			localServerID, peerServer1ID, actor.Term(2), 0, actor.LogIndex(2), actor.Term(1),
			[]actor.LogEntry{
				actor.NewLogEntry(1, 1, ""),
				actor.NewLogEntry(1, 2, ""),
				actor.NewLogEntry(2, 3, ""),
			}))

	assert.Equal(t, actor.LogIndex(4), storage.LatestLogEntry().Index)
}

func TestMsgAppendEntries__AppendingEntriesWithConflictingTermsPrunesOldEntries(t *testing.T) {
	act, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer1ID, actor.NewMessageAppendEntriesResponse(
		peerServer1ID, localServerID, actor.Term(2), true, actor.LogIndex(2)))

	storage.SetCurrentTerm(1)
	storage.AppendLog("") // term 1, index 1
	storage.AppendLog("") // term 1, index 2

	storage.SetCurrentTerm(2)
	storage.AppendLog("") // term 2, index 3
	storage.AppendLog("") // term 2, index 4

	act.Process(
		actor.NewMessageAppendEntries(
			localServerID, peerServer1ID, actor.Term(2), 0, actor.LogIndex(1), actor.Term(1),
			[]actor.LogEntry{
				actor.NewLogEntry(2, 2, ""),
			}))

	assert.Equal(t, actor.LogIndex(2), storage.LatestLogEntry().Index)
}
