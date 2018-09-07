package actor

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
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

	assert.Equal(t, Term(1), storage.CurrentTerm())

	gatewayMock.EXPECT().Send(gomock.Any(), gomock.Any())

	actor.Process(NewMessageAppendEntries(
		localServerID, peerServer1ID, Term(2), 0, 0, 0, nil))

	assert.Equal(t, Term(2), storage.CurrentTerm())
	assert.Equal(t, FollowerMode, actor.Mode())
}

func TestAppendEntries__LeaderRepeatsHeartbeatsWithDecrementingPrevPointUntilFollowersAck(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	storage.AppendLog("") // index 1, term 0
	storage.AppendLog("") // index 2, term 0
	storage.AppendLog("") // index 3, term 0
	storage.AppendLog("") // index 4, term 0

	testTransitionFromFollowerToLeader(actor, gatewayMock, storage)

	for i := 3; i >= 1; i-- {
		// Reject responses should decrease prev point
		gatewayMock.EXPECT().Send(peerServer1ID, NewMessageAppendEntries(
			peerServer1ID, localServerID, Term(1), 0, LogIndex(i), 0, nil))
		gatewayMock.EXPECT().Send(peerServer2ID, NewMessageAppendEntries(
			peerServer2ID, localServerID, Term(1), 0, LogIndex(i), 0, nil))

		actor.Process(NewMessageAppendEntriesResponse(
			localServerID, peerServer1ID, 1, false, 0))
		actor.Process(NewMessageAppendEntriesResponse(
			localServerID, peerServer2ID, 1, false, 0))
	}
}

func TestAppendEntries__MatchAndNextIndexAreUpdatedAccordinglyOnSuccessfulAppendEntriesResponse(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	storage.AppendLog("") // index 1, term 0
	storage.AppendLog("") // index 2, term 0
	storage.AppendLog("") // index 3, term 0
	storage.AppendLog("") // index 4, term 0

	testTransitionFromFollowerToLeader(actor, gatewayMock, storage)

	actor.Process(NewMessageAppendEntriesResponse(
		localServerID, peerServer1ID, 1, true, 4))
	actor.Process(NewMessageAppendEntriesResponse(
		localServerID, peerServer2ID, 1, true, 4))

	gatewayMock.EXPECT().Send(peerServer1ID, NewMessageAppendEntries(
		peerServer1ID, localServerID, Term(1), 4, 4, 0, nil))
	gatewayMock.EXPECT().Send(peerServer2ID, NewMessageAppendEntries(
		peerServer2ID, localServerID, Term(1), 4, 4, 0, nil))

	testProgressTime(actor, 5)
}

func TestTick__SendsHeartbeatPeriodically(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	testTransitionFromFollowerToLeader(actor, gatewayMock, storage)

	for i := 0; i < 3; i++ {
		gatewayMock.EXPECT().Send(peerServer1ID, NewMessageAppendEntries(
			peerServer1ID, localServerID, Term(1), 0, 0, 0, nil))
		gatewayMock.EXPECT().Send(peerServer2ID, NewMessageAppendEntries(
			peerServer2ID, localServerID, Term(1), 0, 0, 0, nil))

		testProgressTime(actor, 5)
	}
}

func TestTick__CommitIndexIsIncreasedWhenChangeIsReplicatedToMajority(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	storage.AppendLog("") // index 1, term 0
	storage.AppendLog("") // index 2, term 0
	storage.AppendLog("") // index 3, term 0
	storage.AppendLog("") // index 4, term 0

	testTransitionFromFollowerToLeader(actor, gatewayMock, storage)

	actor.Process(NewMessageAppendEntriesResponse(
		localServerID, peerServer1ID, 1, true, 4))
	actor.Process(NewMessageAppendEntriesResponse(
		localServerID, peerServer2ID, 1, true, 4))

	expectedCommitIndex := LogIndex(4)

	gatewayMock.EXPECT().Send(peerServer1ID, NewMessageAppendEntries(
		peerServer1ID, localServerID, Term(1), expectedCommitIndex, 4, 0, nil))
	gatewayMock.EXPECT().Send(peerServer2ID, NewMessageAppendEntries(
		peerServer2ID, localServerID, Term(1), expectedCommitIndex, 4, 0, nil))

	testProgressTime(actor, 5)
}

func TestTick__CommitIndexIsNewerIncreasedBeyondMatchIndex(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	storage.AppendLog("") // index 1, term 0
	storage.AppendLog("") // index 2, term 0
	storage.AppendLog("") // index 3, term 0
	storage.AppendLog("") // index 4, term 0

	testTransitionFromFollowerToLeader(actor, gatewayMock, storage)

	actor.Process(NewMessageAppendEntriesResponse(
		localServerID, peerServer1ID, 1, true, 4))
	actor.Process(NewMessageAppendEntriesResponse(
		localServerID, peerServer2ID, 1, true, 3))

	gatewayMock.EXPECT().Send(peerServer1ID, NewMessageAppendEntries(
		peerServer1ID, localServerID, Term(1), 4, 4, 0, nil))
	gatewayMock.EXPECT().Send(peerServer2ID, NewMessageAppendEntries(
		peerServer2ID, localServerID, Term(1), 3, 3, 0, []LogEntry{
			NewLogEntry(0, 4, ""),
		}))

	testProgressTime(actor, 5)
}

func TestProposal__ProposalsAppendToLocalLog(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	storage.AppendLog("") // index 1, term 0
	storage.AppendLog("") // index 2, term 0
	storage.AppendLog("") // index 3, term 0
	storage.AppendLog("") // index 4, term 0

	testTransitionFromFollowerToLeader(actor, gatewayMock, storage)

	actor.Process(NewMessageProposal(localServerID, localServerID, ""))

	assert.Equal(t, LogIndex(5), storage.LatestLogEntry().Index)
}

func TestProposal__ProposalsAreReplicatedOnHeartbeat(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	storage.AppendLog("") // index 1, term 0
	storage.AppendLog("") // index 2, term 0
	storage.AppendLog("") // index 3, term 0
	storage.AppendLog("") // index 4, term 0

	testTransitionFromFollowerToLeader(actor, gatewayMock, storage)
	testFollowersReplicateUp(actor, storage)

	actor.Process(NewMessageProposal(localServerID, localServerID, ""))

	logEntries := []LogEntry{
		NewLogEntry(1, 5, ""),
	}

	gatewayMock.EXPECT().Send(peerServer1ID, NewMessageAppendEntries(
		peerServer1ID, localServerID, Term(1), 4, 4, 0, logEntries))
	gatewayMock.EXPECT().Send(peerServer2ID, NewMessageAppendEntries(
		peerServer2ID, localServerID, Term(1), 4, 4, 0, logEntries))

	testProgressTime(actor, 5)

}
