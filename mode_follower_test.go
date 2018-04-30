package go_raft_test

import (
	"github.com/golang/mock/gomock"
	"github.com/sema/go-raft"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMsgVoteFor__IsAbleToGetAVote(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer1ID, go_raft.NewMessageVoteForResponse(
		peerServer1ID, localServerID, go_raft.Term(1), true))

	actor.Process(
		go_raft.NewMessageVoteFor(localServerID, peerServer1ID, go_raft.Term(1), 0, 0))

	assert.Equal(t, storage.VotedFor(), peerServer1ID)
}

func TestMsgVoteFor__DoesNotOverwriteExistingVoteInTerm(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	storage.SetCurrentTerm(go_raft.Term(1))
	storage.SetVotedForIfUnset(peerServer1ID)

	gatewayMock.EXPECT().Send(peerServer2ID, go_raft.NewMessageVoteForResponse(
		peerServer2ID, localServerID, go_raft.Term(1), false))

	actor.Process(
		go_raft.NewMessageVoteFor(localServerID, peerServer2ID, go_raft.Term(1), 0, 0))

	assert.Equal(t, storage.VotedFor(), peerServer1ID)
}

func TestMsgTick__FollowerEventuallyTransitionsToCandidateAfterXTicks(t *testing.T) {
	actor, gatewayMock, _, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(gomock.Any(), gomock.Any()).AnyTimes()

	testProgressTime(actor, 11)

	assert.Equal(t, go_raft.CandidateMode, actor.Mode())
}

func TestMsgVoteFor__IsRejectedIfIfCandidateLogBelongsToOlderTerm(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer1ID, go_raft.NewMessageVoteForResponse(
		peerServer1ID, localServerID, go_raft.Term(1), false))

	storage.SetCurrentTerm(1)
	storage.AppendLog("") // term 1, index 1
	storage.AppendLog("") // term 1, index 2

	actor.Process(
		go_raft.NewMessageVoteFor(localServerID, peerServer1ID, go_raft.Term(1), 0, 0))
}

func TestMsgVoteFor__IsRejectedIfIfCandidateLogIsBehindOnIndex(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer1ID, go_raft.NewMessageVoteForResponse(
		peerServer1ID, localServerID, go_raft.Term(1), false))

	storage.SetCurrentTerm(1)
	storage.AppendLog("") // term 1, index 1
	storage.AppendLog("") // term 1, index 2

	actor.Process(
		go_raft.NewMessageVoteFor(localServerID, peerServer1ID, go_raft.Term(1), 1, 1))
}

func TestMsgAppendEntries__IsRejectedIfPreviousTermAndIndexDoesNotMatch(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer1ID, go_raft.NewMessageAppendEntriesResponse(
		peerServer1ID, localServerID, go_raft.Term(1), false, go_raft.LogIndex(0)))

	storage.SetCurrentTerm(1)
	storage.AppendLog("") // term 1, index 1
	storage.AppendLog("") // term 1, index 2

	actor.Process(
		go_raft.NewMessageAppendEntries(
			localServerID, peerServer1ID, go_raft.Term(1), 0, go_raft.LogIndex(3), go_raft.Term(3), []go_raft.LogEntry{}))
}

func TestMsgAppendEntries__IsAcceptedIfPreviousTermAndIndexMatch(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer1ID, go_raft.NewMessageAppendEntriesResponse(
		peerServer1ID, localServerID, go_raft.Term(1), true, go_raft.LogIndex(2)))

	storage.SetCurrentTerm(1)
	storage.AppendLog("") // term 1, index 1
	storage.AppendLog("") // term 1, index 2

	actor.Process(
		go_raft.NewMessageAppendEntries(
			localServerID, peerServer1ID, go_raft.Term(1), 0, go_raft.LogIndex(2), go_raft.Term(1), []go_raft.LogEntry{}))
}

func TestMsgAppendEntries__AppendsNewEntriesToLog(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer1ID, go_raft.NewMessageAppendEntriesResponse(
		peerServer1ID, localServerID, go_raft.Term(1), true, go_raft.LogIndex(4)))

	storage.SetCurrentTerm(1)
	storage.AppendLog("") // term 1, index 1
	storage.AppendLog("") // term 1, index 2

	actor.Process(
		go_raft.NewMessageAppendEntries(
			localServerID, peerServer1ID, go_raft.Term(1), 0, go_raft.LogIndex(2), go_raft.Term(1),
			[]go_raft.LogEntry{
				go_raft.NewLogEntry(1, 3, ""),
				go_raft.NewLogEntry(1, 4, ""),
			}))

	assert.Equal(t, go_raft.LogIndex(4), storage.LatestLogEntry().Index)
}

func TestMsgAppendEntries__AppendingPreviouslyAppendedEntriesRetainsCurrentState(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer1ID, go_raft.NewMessageAppendEntriesResponse(
		peerServer1ID, localServerID, go_raft.Term(2), true, go_raft.LogIndex(4)))

	storage.SetCurrentTerm(1)
	storage.AppendLog("") // term 1, index 1
	storage.AppendLog("") // term 1, index 2

	storage.SetCurrentTerm(2)
	storage.AppendLog("") // term 2, index 3
	storage.AppendLog("") // term 2, index 4

	actor.Process(
		go_raft.NewMessageAppendEntries(
			localServerID, peerServer1ID, go_raft.Term(2), 0, go_raft.LogIndex(2), go_raft.Term(1),
			[]go_raft.LogEntry{
				go_raft.NewLogEntry(1, 1, ""),
				go_raft.NewLogEntry(1, 2, ""),
				go_raft.NewLogEntry(2, 3, ""),
			}))

	assert.Equal(t, go_raft.LogIndex(4), storage.LatestLogEntry().Index)
}

func TestMsgAppendEntries__AppendingEntriesWithConflictingTermsPrunesOldEntries(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer1ID, go_raft.NewMessageAppendEntriesResponse(
		peerServer1ID, localServerID, go_raft.Term(2), true, go_raft.LogIndex(2)))

	storage.SetCurrentTerm(1)
	storage.AppendLog("") // term 1, index 1
	storage.AppendLog("") // term 1, index 2

	storage.SetCurrentTerm(2)
	storage.AppendLog("") // term 2, index 3
	storage.AppendLog("") // term 2, index 4

	actor.Process(
		go_raft.NewMessageAppendEntries(
			localServerID, peerServer1ID, go_raft.Term(2), 0, go_raft.LogIndex(1), go_raft.Term(1),
			[]go_raft.LogEntry{
				go_raft.NewLogEntry(2, 2, ""),
			}))

	assert.Equal(t, go_raft.LogIndex(2), storage.LatestLogEntry().Index)
}
