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

// Base test setup for an actor, with in-memory storage, mocked gateway, and 2 peer actors.
func newActorTestSetup(t *testing.T) (
	go_raft.Actor, *mock_go_raft.MockServerGateway, go_raft.PersistentStorage, func()) {

	mockCtrl := gomock.NewController(t)
	cleanup := func() {
		mockCtrl.Finish()
	}

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
	storage.AppendLog()  // term 1, index 1
	storage.AppendLog()  // term 1, index 2

	actor.Process(
		go_raft.NewMessageVoteFor(localServerID, peerServer1ID, go_raft.Term(1), 0, 0))
}

func TestMsgVoteFor__IsRejectedIfIfCandidateLogIsBehindOnIndex(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	gatewayMock.EXPECT().Send(peerServer1ID, go_raft.NewMessageVoteForResponse(
		peerServer1ID, localServerID, go_raft.Term(1), false))

	storage.SetCurrentTerm(1)
	storage.AppendLog()  // term 1, index 1
	storage.AppendLog()  // term 1, index 2

	actor.Process(
		go_raft.NewMessageVoteFor(localServerID, peerServer1ID, go_raft.Term(1), 1, 1))
}

