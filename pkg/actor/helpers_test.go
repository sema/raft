package actor

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sema/raft/pkg/actor/mocks"
)

const localServerID = ServerID("server1.servers.local")
const peerServer1ID = ServerID("server2.servers.local")
const peerServer2ID = ServerID("server3.servers.local")

// Base test setup for an actor, with in-memory storage, mocked gateway, and 2 peer actors.
func newActorTestSetup(t *testing.T) (
	Actor, *mock_actor.MockServerGateway, Storage, func()) {

	mockCtrl := gomock.NewController(t)
	cleanup := mockCtrl.Finish

	gatewayMock := mock_actor.NewMockServerGateway(mockCtrl)
	storage := NewMemoryStorage()

	config := Config{
		Servers:                    []ServerID{localServerID, peerServer1ID, peerServer2ID},
		LeaderElectionTimeout:      10,
		LeaderElectionTimeoutSplay: 0,
		LeaderHeartbeatFrequency:   5,
	}

	actor := NewActor(localServerID, storage, gatewayMock, config)

	return actor, gatewayMock, storage, cleanup
}

func testTransitionFromFollowerToCandidate(actor Actor, gatewayMock *mock_actor.MockServerGateway, storage Storage) {
	prevIndex := storage.LatestLogEntry().Index
	prevTerm := storage.LatestLogEntry().Term

	gatewayMock.EXPECT().Send(peerServer1ID, NewMessageVoteFor(
		peerServer1ID, localServerID, Term(1), prevIndex, prevTerm))
	gatewayMock.EXPECT().Send(peerServer2ID, NewMessageVoteFor(
		peerServer2ID, localServerID, Term(1), prevIndex, prevTerm))

	testProgressTime(actor, 11)
}

func testProgressTime(actor Actor, numTicks int) {
	for i := 0; i < numTicks; i++ {
		actor.Process(NewMessageTick(localServerID, localServerID))
	}
}

func testTransitionFromFollowerToLeader(actor Actor, gatewayMock *mock_actor.MockServerGateway, storage Storage) {
	testTransitionFromFollowerToCandidate(actor, gatewayMock, storage)

	prevIndex := storage.LatestLogEntry().Index
	prevTerm := storage.LatestLogEntry().Term

	gatewayMock.EXPECT().Send(peerServer1ID, NewMessageAppendEntries(
		peerServer1ID, localServerID, Term(1), 0, prevIndex, prevTerm, nil))
	gatewayMock.EXPECT().Send(peerServer2ID, NewMessageAppendEntries(
		peerServer2ID, localServerID, Term(1), 0, prevIndex, prevTerm, nil))

	actor.Process(NewMessageVoteForResponse(
		localServerID, peerServer1ID, Term(1), true))
}

func testFollowersReplicateUp(actor Actor, storage Storage) {
	term := storage.CurrentTerm()
	lastIndex := storage.LatestLogEntry().Index

	actor.Process(NewMessageAppendEntriesResponse(
		localServerID, peerServer1ID, term, true, lastIndex))
	actor.Process(NewMessageAppendEntriesResponse(
		localServerID, peerServer2ID, term, true, lastIndex))
}
