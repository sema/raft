package go_raft_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/sema/go-raft"
	"github.com/sema/go-raft/mocks"
	"github.com/golang/mock/gomock"
)

func TestEnter__LeaderSendsHeartbeatsToAllPeersUponElection(t *testing.T) {
	actor, gatewayMock, _, cleanup := newActorTestSetup(t)
	defer cleanup()

	// Assertions that heartbeats are sent are in test function
	testTransitionFromFollowerToLeader(actor, gatewayMock)
}


func TestAppendEntries__LeaderTransitionsToFollowerIfNewLeaderIsDetected(t *testing.T) {
	actor, gatewayMock, storage, cleanup := newActorTestSetup(t)
	defer cleanup()

	testTransitionFromFollowerToLeader(actor, gatewayMock)

	assert.Equal(t, go_raft.Term(1), storage.CurrentTerm())

	gatewayMock.EXPECT().Send(gomock.Any(), gomock.Any())

	actor.Process(go_raft.NewMessageAppendEntries(
		localServerID, peerServer1ID, go_raft.Term(2), 0, 0, 0, []go_raft.LogEntry{}))

	assert.Equal(t, go_raft.Term(2), storage.CurrentTerm())
	assert.Equal(t, go_raft.FollowerMode, actor.Mode())
}


// TEST heartbeat is repeated until client responds OK
// TEST adding log entry triggers new heartbeat
// TEST heartbeat responses increase index when majority reach new point

// TEST sends heartbeat regularly