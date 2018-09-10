package actor_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sema/raft/pkg/actor"
	"github.com/sema/raft/pkg/actor/mocks"
)

const localServerID = actor.ServerID("server1.servers.local")
const peerServer1ID = actor.ServerID("server2.servers.local")
const peerServer2ID = actor.ServerID("server3.servers.local")

// Base test setup for an actor, with in-memory storage, mocked gateway, and 2 peer actors.
func newActorTestSetup(t *testing.T) (
	actor.Actor, *mock_actor.MockServerGateway, actor.Storage, func()) {

	mockCtrl := gomock.NewController(t)
	cleanup := mockCtrl.Finish

	gatewayMock := mock_actor.NewMockServerGateway(mockCtrl)
	storage := actor.NewMemoryStorage()

	config := actor.Config{
		Servers:                    []actor.ServerID{localServerID, peerServer1ID, peerServer2ID},
		LeaderElectionTimeout:      10,
		LeaderElectionTimeoutSplay: 0,
		LeaderHeartbeatFrequency:   5,
	}

	act := actor.NewActor(localServerID, storage, gatewayMock, config)

	return act, gatewayMock, storage, cleanup
}

func testTransitionFromFollowerToCandidate(act actor.Actor, gatewayMock *mock_actor.MockServerGateway, storage actor.Storage) {
	prevIndex := storage.LatestLogEntry().Index
	prevTerm := storage.LatestLogEntry().Term

	gatewayMock.EXPECT().Send(peerServer1ID, actor.NewMessageVoteFor(
		peerServer1ID, localServerID, actor.Term(1), prevIndex, prevTerm))
	gatewayMock.EXPECT().Send(peerServer2ID, actor.NewMessageVoteFor(
		peerServer2ID, localServerID, actor.Term(1), prevIndex, prevTerm))

	testProgressTime(act, 11)
}

func testProgressTime(act actor.Actor, numTicks int) {
	for i := 0; i < numTicks; i++ {
		act.Process(actor.NewMessageTick(localServerID, localServerID))
	}
}

func testTransitionFromFollowerToLeader(act actor.Actor, gatewayMock *mock_actor.MockServerGateway, storage actor.Storage) {
	testTransitionFromFollowerToCandidate(act, gatewayMock, storage)

	prevIndex := storage.LatestLogEntry().Index
	prevTerm := storage.LatestLogEntry().Term

	gatewayMock.EXPECT().Send(peerServer1ID, actor.NewMessageAppendEntries(
		peerServer1ID, localServerID, actor.Term(1), 0, prevIndex, prevTerm, nil))
	gatewayMock.EXPECT().Send(peerServer2ID, actor.NewMessageAppendEntries(
		peerServer2ID, localServerID, actor.Term(1), 0, prevIndex, prevTerm, nil))

	act.Process(actor.NewMessageVoteForResponse(
		localServerID, peerServer1ID, actor.Term(1), true))
}

func testFollowersReplicateUp(act actor.Actor, storage actor.Storage) {
	term := storage.CurrentTerm()
	lastIndex := storage.LatestLogEntry().Index

	act.Process(actor.NewMessageAppendEntriesResponse(
		localServerID, peerServer1ID, term, true, lastIndex))
	act.Process(actor.NewMessageAppendEntriesResponse(
		localServerID, peerServer2ID, term, true, lastIndex))
}
