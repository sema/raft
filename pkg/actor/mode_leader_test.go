package actor_test

import (
	"testing"

	"github.com/sema/raft/pkg/actor"

	"github.com/stretchr/testify/assert"
)

func TestEnter__LeaderSendsHeartbeatsToAllPeersUponElection(t *testing.T) {
	act, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	// Assertions that heartbeats are sent are in test function
	testTransitionFromFollowerToLeader(t, act, storage)
}

func TestAppendEntries__LeaderTransitionsToFollowerIfNewLeaderIsDetected(t *testing.T) {
	act, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	testTransitionFromFollowerToLeader(t, act, storage)

	assert.Equal(t, actor.Term(1), storage.CurrentTerm())

	act.Process(actor.NewMessageAppendEntries(
		localServerID, peerServer1ID, actor.Term(2), 0, 0, 0, nil))

	assert.Equal(t, actor.Term(2), storage.CurrentTerm())
	assert.Equal(t, actor.FollowerMode, act.Mode())
}

func TestAppendEntries__LeaderRepeatsHeartbeatsWithDecrementingPrevPointUntilFollowersAck(t *testing.T) {
	act, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	storage.AppendLog("") // index 1, term 0
	storage.AppendLog("") // index 2, term 0
	storage.AppendLog("") // index 3, term 0
	storage.AppendLog("") // index 4, term 0

	testTransitionFromFollowerToLeader(t, act, storage)

	for i := 3; i >= 1; i-- {
		// Reject responses should decrease prev point

		messageExpected := []actor.Message{
			actor.NewMessageAppendEntries(peerServer1ID, localServerID, actor.Term(1), 0, actor.LogIndex(i), 0, nil)}
		messageOut := act.Process(actor.NewMessageAppendEntriesResponse(
			localServerID, peerServer1ID, 1, false, 0))
		assert.Equal(t, messageExpected, messageOut)

		messageExpected = []actor.Message{
			actor.NewMessageAppendEntries(peerServer2ID, localServerID, actor.Term(1), 0, actor.LogIndex(i), 0, nil)}
		messageOut = act.Process(actor.NewMessageAppendEntriesResponse(
			localServerID, peerServer2ID, 1, false, 0))
		assert.Equal(t, messageExpected, messageOut)
	}
}

func TestAppendEntries__MatchAndNextIndexAreUpdatedAccordinglyOnSuccessfulAppendEntriesResponse(t *testing.T) {
	act, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	storage.AppendLog("") // index 1, term 0
	storage.AppendLog("") // index 2, term 0
	storage.AppendLog("") // index 3, term 0
	storage.AppendLog("") // index 4, term 0

	testTransitionFromFollowerToLeader(t, act, storage)

	act.Process(actor.NewMessageAppendEntriesResponse(
		localServerID, peerServer1ID, 1, true, 4))
	act.Process(actor.NewMessageAppendEntriesResponse(
		localServerID, peerServer2ID, 1, true, 4))

	messagesExpected := []actor.Message{
		actor.NewMessageAppendEntries(peerServer1ID, localServerID, actor.Term(1), 4, 4, 0, nil),
		actor.NewMessageAppendEntries(peerServer2ID, localServerID, actor.Term(1), 4, 4, 0, nil),
	}
	messagesOut := testProgressTime(act, 5)

	assert.Equal(t, messagesExpected, messagesOut)
}

func TestTick__SendsHeartbeatPeriodically(t *testing.T) {
	act, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	testTransitionFromFollowerToLeader(t, act, storage)

	for i := 0; i < 3; i++ {
		messagesExpected := []actor.Message{
			actor.NewMessageAppendEntries(peerServer1ID, localServerID, actor.Term(1), 0, 0, 0, nil),
			actor.NewMessageAppendEntries(peerServer2ID, localServerID, actor.Term(1), 0, 0, 0, nil),
		}
		messagesOut := testProgressTime(act, 5)
		assert.Equal(t, messagesExpected, messagesOut)
	}
}

func TestTick__CommitIndexIsIncreasedWhenChangeIsReplicatedToMajority(t *testing.T) {
	act, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	storage.AppendLog("") // index 1, term 0
	storage.AppendLog("") // index 2, term 0
	storage.AppendLog("") // index 3, term 0
	storage.AppendLog("") // index 4, term 0

	testTransitionFromFollowerToLeader(t, act, storage)

	act.Process(actor.NewMessageAppendEntriesResponse(
		localServerID, peerServer1ID, 1, true, 4))
	act.Process(actor.NewMessageAppendEntriesResponse(
		localServerID, peerServer2ID, 1, true, 4))

	expectedCommitIndex := actor.LogIndex(4)

	messagesExpected := []actor.Message{
		actor.NewMessageAppendEntries(peerServer1ID, localServerID, actor.Term(1), expectedCommitIndex, 4, 0, nil),
		actor.NewMessageAppendEntries(peerServer2ID, localServerID, actor.Term(1), expectedCommitIndex, 4, 0, nil),
	}
	messagesOut := testProgressTime(act, 5)
	assert.Equal(t, messagesExpected, messagesOut)
}

func TestTick__CommitIndexIsNewerIncreasedBeyondMatchIndex(t *testing.T) {
	act, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	storage.AppendLog("") // index 1, term 0
	storage.AppendLog("") // index 2, term 0
	storage.AppendLog("") // index 3, term 0
	storage.AppendLog("") // index 4, term 0

	testTransitionFromFollowerToLeader(t, act, storage)

	act.Process(actor.NewMessageAppendEntriesResponse(
		localServerID, peerServer1ID, 1, true, 4))
	act.Process(actor.NewMessageAppendEntriesResponse(
		localServerID, peerServer2ID, 1, true, 3))

	messagesExpected := []actor.Message{
		actor.NewMessageAppendEntries(peerServer1ID, localServerID, actor.Term(1), 4, 4, 0, nil),
		actor.NewMessageAppendEntries(peerServer2ID, localServerID, actor.Term(1), 3, 3, 0,
			[]actor.LogEntry{actor.NewLogEntry(0, 4, "")}),
	}

	messagesOut := testProgressTime(act, 5)
	assert.Equal(t, messagesExpected, messagesOut)
}

func TestProposal__ProposalsAppendToLocalLog(t *testing.T) {
	act, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	storage.AppendLog("") // index 1, term 0
	storage.AppendLog("") // index 2, term 0
	storage.AppendLog("") // index 3, term 0
	storage.AppendLog("") // index 4, term 0

	testTransitionFromFollowerToLeader(t, act, storage)

	act.Process(actor.NewMessageProposal(localServerID, localServerID, ""))

	assert.Equal(t, actor.LogIndex(5), storage.LatestLogEntry().Index)
}

func TestProposal__ProposalsAreReplicatedOnHeartbeat(t *testing.T) {
	act, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	storage.AppendLog("") // index 1, term 0
	storage.AppendLog("") // index 2, term 0
	storage.AppendLog("") // index 3, term 0
	storage.AppendLog("") // index 4, term 0

	testTransitionFromFollowerToLeader(t, act, storage)
	testFollowersReplicateUp(act, storage)

	act.Process(actor.NewMessageProposal(localServerID, localServerID, ""))

	logEntries := []actor.LogEntry{
		actor.NewLogEntry(1, 5, ""),
	}

	messagesExpected := []actor.Message{
		actor.NewMessageAppendEntries(peerServer1ID, localServerID, actor.Term(1), 4, 4, 0, logEntries),
		actor.NewMessageAppendEntries(peerServer2ID, localServerID, actor.Term(1), 4, 4, 0, logEntries),
	}
	messagesOut := testProgressTime(act, 5)
	assert.Equal(t, messagesExpected, messagesOut)

}
