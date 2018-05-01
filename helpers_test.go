package raft_test

import (
	"github.com/golang/mock/gomock"
	"github.com/sema/go-raft"
	"github.com/sema/go-raft/mocks"
	"testing"
)

const localServerID = raft.ServerID("server1.servers.local")
const peerServer1ID = raft.ServerID("server2.servers.local")
const peerServer2ID = raft.ServerID("server3.servers.local")

// Base test setup for an actor, with in-memory storage, mocked gateway, and 2 peer actors.
func newActorTestSetup(t *testing.T) (
	raft.Actor, *mock_raft.MockServerGateway, raft.PersistentStorage, func()) {

	mockCtrl := gomock.NewController(t)
	cleanup := mockCtrl.Finish

	gatewayMock := mock_raft.NewMockServerGateway(mockCtrl)
	storage := raft.NewMemoryStorage()
	discovery := raft.NewStaticDiscovery(
		[]raft.ServerID{localServerID, peerServer1ID, peerServer2ID})

	config := raft.Config{
		LeaderElectionTimeout:      10,
		LeaderElectionTimeoutSplay: 0,
		LeaderHeartbeatFrequency:   5,
	}

	actor := raft.NewActor(localServerID, storage, gatewayMock, discovery, config)

	return actor, gatewayMock, storage, cleanup
}

func testTransitionFromFollowerToCandidate(actor raft.Actor, gatewayMock *mock_raft.MockServerGateway, storage raft.PersistentStorage) {
	prevIndex := storage.LatestLogEntry().Index
	prevTerm := storage.LatestLogEntry().Term

	gatewayMock.EXPECT().Send(peerServer1ID, raft.NewMessageVoteFor(
		peerServer1ID, localServerID, raft.Term(1), prevIndex, prevTerm))
	gatewayMock.EXPECT().Send(peerServer2ID, raft.NewMessageVoteFor(
		peerServer2ID, localServerID, raft.Term(1), prevIndex, prevTerm))

	testProgressTime(actor, 11)
}

func testProgressTime(actor raft.Actor, numTicks int) {
	for i := 0; i < numTicks; i++ {
		actor.Process(raft.NewMessageTick(localServerID, localServerID))
	}
}

func testTransitionFromFollowerToLeader(actor raft.Actor, gatewayMock *mock_raft.MockServerGateway, storage raft.PersistentStorage) {
	testTransitionFromFollowerToCandidate(actor, gatewayMock, storage)

	prevIndex := storage.LatestLogEntry().Index
	prevTerm := storage.LatestLogEntry().Term

	gatewayMock.EXPECT().Send(peerServer1ID, raft.NewMessageAppendEntries(
		peerServer1ID, localServerID, raft.Term(1), 0, prevIndex, prevTerm, nil))
	gatewayMock.EXPECT().Send(peerServer2ID, raft.NewMessageAppendEntries(
		peerServer2ID, localServerID, raft.Term(1), 0, prevIndex, prevTerm, nil))

	actor.Process(raft.NewMessageVoteForResponse(
		localServerID, peerServer1ID, raft.Term(1), true))
}

func testFollowersReplicateUp(actor raft.Actor, storage raft.PersistentStorage) {
	term := storage.CurrentTerm()
	lastIndex := storage.LatestLogEntry().Index

	actor.Process(raft.NewMessageAppendEntriesResponse(
		localServerID, peerServer1ID, term, true, lastIndex))
	actor.Process(raft.NewMessageAppendEntriesResponse(
		localServerID, peerServer2ID, term, true, lastIndex))
}
