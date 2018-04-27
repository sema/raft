package go_raft_test

import (
	"github.com/golang/mock/gomock"
	"github.com/sema/go-raft"
	"github.com/sema/go-raft/mocks"
	"testing"
	"github.com/stretchr/testify/assert"
)

const localServerID = go_raft.ServerID("server1.servers.local")
const peerServer1ID = go_raft.ServerID("server2.servers.local")
const peerServer2ID = go_raft.ServerID("server3.servers.local")

// TODO move test helper methods somewhere else
// Base test setup for an actor, with in-memory storage, mocked gateway, and 2 peer actors.
func newActorTestSetup(t *testing.T) (
	go_raft.Actor, *mock_go_raft.MockServerGateway, go_raft.PersistentStorage, func()) {

	mockCtrl := gomock.NewController(t)
	cleanup := mockCtrl.Finish

	gatewayMock := mock_go_raft.NewMockServerGateway(mockCtrl)
	storage := go_raft.NewMemoryStorage()
	discovery := go_raft.NewStaticDiscovery(
		[]go_raft.ServerID{localServerID, peerServer1ID, peerServer2ID})

	actor := go_raft.NewActor(localServerID, storage, gatewayMock, discovery)

	return actor, gatewayMock, storage, cleanup
}

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
	storage.AppendLog("")  // term 1, index 1
	storage.AppendLog("")  // term 1, index 2

	actor.Process(
		go_raft.NewMessageVoteFor(localServerID, peerServer1ID, go_raft.Term(1), 0, 0))
}

func TestMsgVoteFor__IsRejectedIfIfCandidateLogIsBehindOnIndex(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer1ID, go_raft.NewMessageVoteForResponse(
		peerServer1ID, localServerID, go_raft.Term(1), false))

	storage.SetCurrentTerm(1)
	storage.AppendLog("")  // term 1, index 1
	storage.AppendLog("")  // term 1, index 2

	actor.Process(
		go_raft.NewMessageVoteFor(localServerID, peerServer1ID, go_raft.Term(1), 1, 1))
}

func TestMsgAppendEntries__IsRejectedIfPreviousTermAndIndexDoesNotMatch(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer1ID, go_raft.NewMessageAppendEntriesResponse(
		peerServer1ID, localServerID, go_raft.Term(1), false, go_raft.LogIndex(0)))

	storage.SetCurrentTerm(1)
	storage.AppendLog("")  // term 1, index 1
	storage.AppendLog("")  // term 1, index 2

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
	storage.AppendLog("")  // term 1, index 1
	storage.AppendLog("")  // term 1, index 2

	actor.Process(
		go_raft.NewMessageAppendEntries(
			localServerID, peerServer1ID, go_raft.Term(1), 0, go_raft.LogIndex(2), go_raft.Term(1), []go_raft.LogEntry{}))
}


func TestMsgAppendEntries__PrunesOldEntriesIfAfterAgreedPreviousPoint(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer1ID, go_raft.NewMessageAppendEntriesResponse(
		peerServer1ID, localServerID, go_raft.Term(3), true, go_raft.LogIndex(2)))

	storage.SetCurrentTerm(1)
	storage.AppendLog("")  // term 1, index 1
	storage.AppendLog("")  // term 1, index 2

	storage.SetCurrentTerm(2)
	storage.AppendLog("")  // term 2, index 3
	storage.AppendLog("")  // term 2, index 4

	actor.Process(
		go_raft.NewMessageAppendEntries(
			localServerID, peerServer1ID, go_raft.Term(3), 0, go_raft.LogIndex(2), go_raft.Term(1), []go_raft.LogEntry{}))

	assert.Equal(t, go_raft.LogIndex(2), storage.LatestLogEntry().Index)
}

func TestMsgAppendEntries__AppendsNewEntriesToLog(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer1ID, go_raft.NewMessageAppendEntriesResponse(
		peerServer1ID, localServerID, go_raft.Term(1), true, go_raft.LogIndex(4)))

	storage.SetCurrentTerm(1)
	storage.AppendLog("")  // term 1, index 1
	storage.AppendLog("")  // term 1, index 2

	actor.Process(
		go_raft.NewMessageAppendEntries(
			localServerID, peerServer1ID, go_raft.Term(1), 0, go_raft.LogIndex(2), go_raft.Term(1),
			[]go_raft.LogEntry{
				go_raft.NewLogEntry(1, 3, ""),
				go_raft.NewLogEntry(1, 4, ""),
			}))

	assert.Equal(t, go_raft.LogIndex(4), storage.LatestLogEntry().Index)
}

// TODO figure out what happens if we have out-of-order heartbeats, and we process and old heartbeat which would then
// prune entries?