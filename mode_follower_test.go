package raft_test

import (
	"github.com/golang/mock/gomock"
	"github.com/sema/go-raft"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMsgVoteFor__IsAbleToGetAVote(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer1ID, raft.NewMessageVoteForResponse(
		peerServer1ID, localServerID, raft.Term(1), true))

	actor.Process(
		raft.NewMessageVoteFor(localServerID, peerServer1ID, raft.Term(1), 0, 0))

	assert.Equal(t, storage.VotedFor(), peerServer1ID)
}

func TestMsgVoteFor__DoesNotOverwriteExistingVoteInTerm(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	storage.SetCurrentTerm(raft.Term(1))
	storage.SetVotedForIfUnset(peerServer1ID)

	gatewayMock.EXPECT().Send(peerServer2ID, raft.NewMessageVoteForResponse(
		peerServer2ID, localServerID, raft.Term(1), false))

	actor.Process(
		raft.NewMessageVoteFor(localServerID, peerServer2ID, raft.Term(1), 0, 0))

	assert.Equal(t, storage.VotedFor(), peerServer1ID)
}

func TestMsgTick__FollowerEventuallyTransitionsToCandidateAfterXTicks(t *testing.T) {
	actor, gatewayMock, _, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(gomock.Any(), gomock.Any()).AnyTimes()

	testProgressTime(actor, 11)

	assert.Equal(t, raft.CandidateMode, actor.Mode())
}

func TestMsgVoteFor__IsRejectedIfIfCandidateLogBelongsToOlderTerm(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer1ID, raft.NewMessageVoteForResponse(
		peerServer1ID, localServerID, raft.Term(1), false))

	storage.SetCurrentTerm(1)
	storage.AppendLog("") // term 1, index 1
	storage.AppendLog("") // term 1, index 2

	actor.Process(
		raft.NewMessageVoteFor(localServerID, peerServer1ID, raft.Term(1), 0, 0))
}

func TestMsgVoteFor__IsRejectedIfIfCandidateLogIsBehindOnIndex(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer1ID, raft.NewMessageVoteForResponse(
		peerServer1ID, localServerID, raft.Term(1), false))

	storage.SetCurrentTerm(1)
	storage.AppendLog("") // term 1, index 1
	storage.AppendLog("") // term 1, index 2

	actor.Process(
		raft.NewMessageVoteFor(localServerID, peerServer1ID, raft.Term(1), 1, 1))
}

func TestMsgAppendEntries__IsRejectedIfPreviousTermAndIndexDoesNotMatch(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer1ID, raft.NewMessageAppendEntriesResponse(
		peerServer1ID, localServerID, raft.Term(1), false, raft.LogIndex(0)))

	storage.SetCurrentTerm(1)
	storage.AppendLog("") // term 1, index 1
	storage.AppendLog("") // term 1, index 2

	actor.Process(
		raft.NewMessageAppendEntries(
			localServerID, peerServer1ID, raft.Term(1), 0, raft.LogIndex(3), raft.Term(3), []raft.LogEntry{}))
}

func TestMsgAppendEntries__IsAcceptedIfPreviousTermAndIndexMatch(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer1ID, raft.NewMessageAppendEntriesResponse(
		peerServer1ID, localServerID, raft.Term(1), true, raft.LogIndex(2)))

	storage.SetCurrentTerm(1)
	storage.AppendLog("") // term 1, index 1
	storage.AppendLog("") // term 1, index 2

	actor.Process(
		raft.NewMessageAppendEntries(
			localServerID, peerServer1ID, raft.Term(1), 0, raft.LogIndex(2), raft.Term(1), []raft.LogEntry{}))
}

func TestMsgAppendEntries__AppendsNewEntriesToLog(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer1ID, raft.NewMessageAppendEntriesResponse(
		peerServer1ID, localServerID, raft.Term(1), true, raft.LogIndex(4)))

	storage.SetCurrentTerm(1)
	storage.AppendLog("") // term 1, index 1
	storage.AppendLog("") // term 1, index 2

	actor.Process(
		raft.NewMessageAppendEntries(
			localServerID, peerServer1ID, raft.Term(1), 0, raft.LogIndex(2), raft.Term(1),
			[]raft.LogEntry{
				raft.NewLogEntry(1, 3, ""),
				raft.NewLogEntry(1, 4, ""),
			}))

	assert.Equal(t, raft.LogIndex(4), storage.LatestLogEntry().Index)
}

func TestMsgAppendEntries__AppendingPreviouslyAppendedEntriesRetainsCurrentState(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer1ID, raft.NewMessageAppendEntriesResponse(
		peerServer1ID, localServerID, raft.Term(2), true, raft.LogIndex(4)))

	storage.SetCurrentTerm(1)
	storage.AppendLog("") // term 1, index 1
	storage.AppendLog("") // term 1, index 2

	storage.SetCurrentTerm(2)
	storage.AppendLog("") // term 2, index 3
	storage.AppendLog("") // term 2, index 4

	actor.Process(
		raft.NewMessageAppendEntries(
			localServerID, peerServer1ID, raft.Term(2), 0, raft.LogIndex(2), raft.Term(1),
			[]raft.LogEntry{
				raft.NewLogEntry(1, 1, ""),
				raft.NewLogEntry(1, 2, ""),
				raft.NewLogEntry(2, 3, ""),
			}))

	assert.Equal(t, raft.LogIndex(4), storage.LatestLogEntry().Index)
}

func TestMsgAppendEntries__AppendingEntriesWithConflictingTermsPrunesOldEntries(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer1ID, raft.NewMessageAppendEntriesResponse(
		peerServer1ID, localServerID, raft.Term(2), true, raft.LogIndex(2)))

	storage.SetCurrentTerm(1)
	storage.AppendLog("") // term 1, index 1
	storage.AppendLog("") // term 1, index 2

	storage.SetCurrentTerm(2)
	storage.AppendLog("") // term 2, index 3
	storage.AppendLog("") // term 2, index 4

	actor.Process(
		raft.NewMessageAppendEntries(
			localServerID, peerServer1ID, raft.Term(2), 0, raft.LogIndex(1), raft.Term(1),
			[]raft.LogEntry{
				raft.NewLogEntry(2, 2, ""),
			}))

	assert.Equal(t, raft.LogIndex(2), storage.LatestLogEntry().Index)
}
