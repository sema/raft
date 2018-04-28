package go_raft_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/sema/go-raft"
	"github.com/golang/mock/gomock"
)

func TestEnter__LeaderSendsHeartbeatsToAllPeersUponElection(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	// Assertions that heartbeats are sent are in test function
	testTransitionFromFollowerToLeader(actor, gatewayMock, storage)
}

func TestAppendEntries__LeaderTransitionsToFollowerIfNewLeaderIsDetected(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	testTransitionFromFollowerToLeader(actor, gatewayMock, storage)

	assert.Equal(t, go_raft.Term(1), storage.CurrentTerm())

	gatewayMock.EXPECT().Send(gomock.Any(), gomock.Any())

	actor.Process(go_raft.NewMessageAppendEntries(
		localServerID, peerServer1ID, go_raft.Term(2), 0, 0, 0, []go_raft.LogEntry{}))

	assert.Equal(t, go_raft.Term(2), storage.CurrentTerm())
	assert.Equal(t, go_raft.FollowerMode, actor.Mode())
}

func TestAppendEntries__LeaderRepeatsHeartbeatsWithDecrementingPrevPointUntilFollowersAck(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	storage.AppendLog("")  // index 1, term 0
	storage.AppendLog("")  // index 2, term 0
	storage.AppendLog("")  // index 3, term 0
	storage.AppendLog("")  // index 4, term 0

	testTransitionFromFollowerToLeader(actor, gatewayMock, storage)

	for i := 3; i >= 1; i-- {
		// Reject responses should decrease prev point
		gatewayMock.EXPECT().Send(peerServer1ID, go_raft.NewMessageAppendEntries(
			peerServer1ID, localServerID, go_raft.Term(1), 0, go_raft.LogIndex(i), 0, []go_raft.LogEntry{}))
		gatewayMock.EXPECT().Send(peerServer2ID, go_raft.NewMessageAppendEntries(
			peerServer2ID, localServerID, go_raft.Term(1), 0, go_raft.LogIndex(i), 0, []go_raft.LogEntry{}))

		actor.Process(go_raft.NewMessageAppendEntriesResponse(
			localServerID, peerServer1ID, 1, false, 0))
		actor.Process(go_raft.NewMessageAppendEntriesResponse(
			localServerID, peerServer2ID, 1, false, 0))
	}
}

func TestAppendEntries__MatchAndNextIndexAreUpdatedAccordinglyOnSuccessfulAppendEntriesResponse(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	storage.AppendLog("")  // index 1, term 0
	storage.AppendLog("")  // index 2, term 0
	storage.AppendLog("")  // index 3, term 0
	storage.AppendLog("")  // index 4, term 0

	testTransitionFromFollowerToLeader(actor, gatewayMock, storage)

	actor.Process(go_raft.NewMessageAppendEntriesResponse(
		localServerID, peerServer1ID, 1, true, 4))
	actor.Process(go_raft.NewMessageAppendEntriesResponse(
		localServerID, peerServer2ID, 1, true, 4))

	gatewayMock.EXPECT().Send(peerServer1ID, go_raft.NewMessageAppendEntries(
		peerServer1ID, localServerID, go_raft.Term(1), 4, 4, 0, []go_raft.LogEntry{}))
	gatewayMock.EXPECT().Send(peerServer2ID, go_raft.NewMessageAppendEntries(
		peerServer2ID, localServerID, go_raft.Term(1), 4, 4, 0, []go_raft.LogEntry{}))

	testProgressTime(actor, 5)
}

func TestTick__SendsHeartbeatPeriodically(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	testTransitionFromFollowerToLeader(actor, gatewayMock, storage)

	for i := 0; i < 3; i++ {
		gatewayMock.EXPECT().Send(peerServer1ID, go_raft.NewMessageAppendEntries(
			peerServer1ID, localServerID, go_raft.Term(1), 0, 0, 0, []go_raft.LogEntry{}))
		gatewayMock.EXPECT().Send(peerServer2ID, go_raft.NewMessageAppendEntries(
			peerServer2ID, localServerID, go_raft.Term(1), 0, 0, 0, []go_raft.LogEntry{}))

		testProgressTime(actor, 5)
	}
}

func TestTick__CommitIndexIsIncreasedWhenChangeIsReplicatedToMajority(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	storage.AppendLog("")  // index 1, term 0
	storage.AppendLog("")  // index 2, term 0
	storage.AppendLog("")  // index 3, term 0
	storage.AppendLog("")  // index 4, term 0

	testTransitionFromFollowerToLeader(actor, gatewayMock, storage)

	actor.Process(go_raft.NewMessageAppendEntriesResponse(
		localServerID, peerServer1ID, 1, true, 4))
	actor.Process(go_raft.NewMessageAppendEntriesResponse(
		localServerID, peerServer2ID, 1, true, 4))

	expectedCommitIndex := go_raft.LogIndex(4)

	gatewayMock.EXPECT().Send(peerServer1ID, go_raft.NewMessageAppendEntries(
		peerServer1ID, localServerID, go_raft.Term(1), expectedCommitIndex, 4, 0, []go_raft.LogEntry{}))
	gatewayMock.EXPECT().Send(peerServer2ID, go_raft.NewMessageAppendEntries(
		peerServer2ID, localServerID, go_raft.Term(1), expectedCommitIndex, 4, 0, []go_raft.LogEntry{}))

	testProgressTime(actor, 5)
}

func TestTick__CommitIndexIsNewerIncreasedBeyondMatchIndex(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	storage.AppendLog("")  // index 1, term 0
	storage.AppendLog("")  // index 2, term 0
	storage.AppendLog("")  // index 3, term 0
	storage.AppendLog("")  // index 4, term 0

	testTransitionFromFollowerToLeader(actor, gatewayMock, storage)

	actor.Process(go_raft.NewMessageAppendEntriesResponse(
		localServerID, peerServer1ID, 1, true, 4))
	actor.Process(go_raft.NewMessageAppendEntriesResponse(
		localServerID, peerServer2ID, 1, true, 3))

	gatewayMock.EXPECT().Send(peerServer1ID, go_raft.NewMessageAppendEntries(
		peerServer1ID, localServerID, go_raft.Term(1), 4, 4, 0, []go_raft.LogEntry{}))
	gatewayMock.EXPECT().Send(peerServer2ID, go_raft.NewMessageAppendEntries(
		peerServer2ID, localServerID, go_raft.Term(1), 3, 3, 0, []go_raft.LogEntry{}))

	testProgressTime(actor, 5)
}

// TEST adding log entry triggers new heartbeat
