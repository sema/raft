package go_raft_test

import (
	"github.com/golang/mock/gomock"
	"github.com/sema/go-raft"
	"github.com/sema/go-raft/mocks"
	"testing"
)

const localServerID = go_raft.ServerID("server1.servers.local")
const peerServer1ID = go_raft.ServerID("server2.servers.local")
const peerServer2ID = go_raft.ServerID("server3.servers.local")

// Base test setup for an actor, with in-memory storage, mocked gateway, and 2 peer actors.
func newActorTestSetup(t *testing.T) (
	go_raft.Actor, *mock_go_raft.MockServerGateway, go_raft.PersistentStorage, func()) {

	mockCtrl := gomock.NewController(t)
	cleanup := mockCtrl.Finish

	gatewayMock := mock_go_raft.NewMockServerGateway(mockCtrl)
	storage := go_raft.NewMemoryStorage()
	discovery := go_raft.NewStaticDiscovery(
		[]go_raft.ServerID{localServerID, peerServer1ID, peerServer2ID})

	config := go_raft.Config{
		LeaderElectionTimeout:      10,
		LeaderElectionTimeoutSplay: 0,
		LeaderHeartbeatFrequency:   5,
	}

	actor := go_raft.NewActor(localServerID, storage, gatewayMock, discovery, config)

	return actor, gatewayMock, storage, cleanup
}

func testTransitionFromFollowerToCandidate(actor go_raft.Actor, gatewayMock *mock_go_raft.MockServerGateway, storage go_raft.PersistentStorage) {
	prevIndex := storage.LatestLogEntry().Index
	prevTerm := storage.LatestLogEntry().Term

	gatewayMock.EXPECT().Send(peerServer1ID, go_raft.NewMessageVoteFor(
		peerServer1ID, localServerID, go_raft.Term(1), prevIndex, prevTerm))
	gatewayMock.EXPECT().Send(peerServer2ID, go_raft.NewMessageVoteFor(
		peerServer2ID, localServerID, go_raft.Term(1), prevIndex, prevTerm))

	testProgressTime(actor, 11)
}

func testProgressTime(actor go_raft.Actor, numTicks int) {
	for i := 0; i < numTicks; i++ {
		actor.Process(go_raft.NewMessageTick(localServerID, localServerID))
	}
}

func testTransitionFromFollowerToLeader(actor go_raft.Actor, gatewayMock *mock_go_raft.MockServerGateway, storage go_raft.PersistentStorage) {
	testTransitionFromFollowerToCandidate(actor, gatewayMock, storage)

	prevIndex := storage.LatestLogEntry().Index
	prevTerm := storage.LatestLogEntry().Term

	gatewayMock.EXPECT().Send(peerServer1ID, go_raft.NewMessageAppendEntries(
		peerServer1ID, localServerID, go_raft.Term(1), 0, prevIndex, prevTerm, nil))
	gatewayMock.EXPECT().Send(peerServer2ID, go_raft.NewMessageAppendEntries(
		peerServer2ID, localServerID, go_raft.Term(1), 0, prevIndex, prevTerm, nil))

	actor.Process(go_raft.NewMessageVoteForResponse(
		localServerID, peerServer1ID, go_raft.Term(1), true))
}

func testFollowersReplicateUp(actor go_raft.Actor, storage go_raft.PersistentStorage) {
	term := storage.CurrentTerm()
	lastIndex := storage.LatestLogEntry().Index

	actor.Process(go_raft.NewMessageAppendEntriesResponse(
		localServerID, peerServer1ID, term, true, lastIndex))
	actor.Process(go_raft.NewMessageAppendEntriesResponse(
		localServerID, peerServer2ID, term, true, lastIndex))
}
