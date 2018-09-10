package actor_test

import (
	"testing"

	"github.com/sema/raft/pkg/actor"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestEnter__LeaderSendsHeartbeatsToAllPeersUponElection(t *testing.T) {
	act, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	// Assertions that heartbeats are sent are in test function
	testTransitionFromFollowerToLeader(act, gatewayMock, storage)
}

func TestAppendEntries__LeaderTransitionsToFollowerIfNewLeaderIsDetected(t *testing.T) {
	act, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	testTransitionFromFollowerToLeader(act, gatewayMock, storage)

	assert.Equal(t, actor.Term(1), storage.CurrentTerm())

	gatewayMock.EXPECT().Send(gomock.Any(), gomock.Any())

	act.Process(actor.NewMessageAppendEntries(
		localServerID, peerServer1ID, actor.Term(2), 0, 0, 0, nil))

	assert.Equal(t, actor.Term(2), storage.CurrentTerm())
	assert.Equal(t, actor.FollowerMode, act.Mode())
}

func TestAppendEntries__LeaderRepeatsHeartbeatsWithDecrementingPrevPointUntilFollowersAck(t *testing.T) {
	act, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	storage.AppendLog("") // index 1, term 0
	storage.AppendLog("") // index 2, term 0
	storage.AppendLog("") // index 3, term 0
	storage.AppendLog("") // index 4, term 0

	testTransitionFromFollowerToLeader(act, gatewayMock, storage)

	for i := 3; i >= 1; i-- {
		// Reject responses should decrease prev point
		gatewayMock.EXPECT().Send(peerServer1ID, actor.NewMessageAppendEntries(
			peerServer1ID, localServerID, actor.Term(1), 0, actor.LogIndex(i), 0, nil))
		gatewayMock.EXPECT().Send(peerServer2ID, actor.NewMessageAppendEntries(
			peerServer2ID, localServerID, actor.Term(1), 0, actor.LogIndex(i), 0, nil))

		act.Process(actor.NewMessageAppendEntriesResponse(
			localServerID, peerServer1ID, 1, false, 0))
		act.Process(actor.NewMessageAppendEntriesResponse(
			localServerID, peerServer2ID, 1, false, 0))
	}
}

func TestAppendEntries__MatchAndNextIndexAreUpdatedAccordinglyOnSuccessfulAppendEntriesResponse(t *testing.T) {
	act, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	storage.AppendLog("") // index 1, term 0
	storage.AppendLog("") // index 2, term 0
	storage.AppendLog("") // index 3, term 0
	storage.AppendLog("") // index 4, term 0

	testTransitionFromFollowerToLeader(act, gatewayMock, storage)

	act.Process(actor.NewMessageAppendEntriesResponse(
		localServerID, peerServer1ID, 1, true, 4))
	act.Process(actor.NewMessageAppendEntriesResponse(
		localServerID, peerServer2ID, 1, true, 4))

	gatewayMock.EXPECT().Send(peerServer1ID, actor.NewMessageAppendEntries(
		peerServer1ID, localServerID, actor.Term(1), 4, 4, 0, nil))
	gatewayMock.EXPECT().Send(peerServer2ID, actor.NewMessageAppendEntries(
		peerServer2ID, localServerID, actor.Term(1), 4, 4, 0, nil))

	testProgressTime(act, 5)
}

func TestTick__SendsHeartbeatPeriodically(t *testing.T) {
	act, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	testTransitionFromFollowerToLeader(act, gatewayMock, storage)

	for i := 0; i < 3; i++ {
		gatewayMock.EXPECT().Send(peerServer1ID, actor.NewMessageAppendEntries(
			peerServer1ID, localServerID, actor.Term(1), 0, 0, 0, nil))
		gatewayMock.EXPECT().Send(peerServer2ID, actor.NewMessageAppendEntries(
			peerServer2ID, localServerID, actor.Term(1), 0, 0, 0, nil))

		testProgressTime(act, 5)
	}
}

func TestTick__CommitIndexIsIncreasedWhenChangeIsReplicatedToMajority(t *testing.T) {
	act, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	storage.AppendLog("") // index 1, term 0
	storage.AppendLog("") // index 2, term 0
	storage.AppendLog("") // index 3, term 0
	storage.AppendLog("") // index 4, term 0

	testTransitionFromFollowerToLeader(act, gatewayMock, storage)

	act.Process(actor.NewMessageAppendEntriesResponse(
		localServerID, peerServer1ID, 1, true, 4))
	act.Process(actor.NewMessageAppendEntriesResponse(
		localServerID, peerServer2ID, 1, true, 4))

	expectedCommitIndex := actor.LogIndex(4)

	gatewayMock.EXPECT().Send(peerServer1ID, actor.NewMessageAppendEntries(
		peerServer1ID, localServerID, actor.Term(1), expectedCommitIndex, 4, 0, nil))
	gatewayMock.EXPECT().Send(peerServer2ID, actor.NewMessageAppendEntries(
		peerServer2ID, localServerID, actor.Term(1), expectedCommitIndex, 4, 0, nil))

	testProgressTime(act, 5)
}

func TestTick__CommitIndexIsNewerIncreasedBeyondMatchIndex(t *testing.T) {
	act, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	storage.AppendLog("") // index 1, term 0
	storage.AppendLog("") // index 2, term 0
	storage.AppendLog("") // index 3, term 0
	storage.AppendLog("") // index 4, term 0

	testTransitionFromFollowerToLeader(act, gatewayMock, storage)

	act.Process(actor.NewMessageAppendEntriesResponse(
		localServerID, peerServer1ID, 1, true, 4))
	act.Process(actor.NewMessageAppendEntriesResponse(
		localServerID, peerServer2ID, 1, true, 3))

	gatewayMock.EXPECT().Send(peerServer1ID, actor.NewMessageAppendEntries(
		peerServer1ID, localServerID, actor.Term(1), 4, 4, 0, nil))
	gatewayMock.EXPECT().Send(peerServer2ID, actor.NewMessageAppendEntries(
		peerServer2ID, localServerID, actor.Term(1), 3, 3, 0, []actor.LogEntry{
			actor.NewLogEntry(0, 4, ""),
		}))

	testProgressTime(act, 5)
}

func TestProposal__ProposalsAppendToLocalLog(t *testing.T) {
	act, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	storage.AppendLog("") // index 1, term 0
	storage.AppendLog("") // index 2, term 0
	storage.AppendLog("") // index 3, term 0
	storage.AppendLog("") // index 4, term 0

	testTransitionFromFollowerToLeader(act, gatewayMock, storage)

	act.Process(actor.NewMessageProposal(localServerID, localServerID, ""))

	assert.Equal(t, actor.LogIndex(5), storage.LatestLogEntry().Index)
}

func TestProposal__ProposalsAreReplicatedOnHeartbeat(t *testing.T) {
	act, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	storage.AppendLog("") // index 1, term 0
	storage.AppendLog("") // index 2, term 0
	storage.AppendLog("") // index 3, term 0
	storage.AppendLog("") // index 4, term 0

	testTransitionFromFollowerToLeader(act, gatewayMock, storage)
	testFollowersReplicateUp(act, storage)

	act.Process(actor.NewMessageProposal(localServerID, localServerID, ""))

	logEntries := []actor.LogEntry{
		actor.NewLogEntry(1, 5, ""),
	}

	gatewayMock.EXPECT().Send(peerServer1ID, actor.NewMessageAppendEntries(
		peerServer1ID, localServerID, actor.Term(1), 4, 4, 0, logEntries))
	gatewayMock.EXPECT().Send(peerServer2ID, actor.NewMessageAppendEntries(
		peerServer2ID, localServerID, actor.Term(1), 4, 4, 0, logEntries))

	testProgressTime(act, 5)

}
