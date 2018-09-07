package raft

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sema/raft/pkg/actor"
	"github.com/sema/raft/pkg/actor/mocks"
)

const localServerID = actor.ServerID("server1.servers.local")

func TestServer_StopCausesStartToReturn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	gatewayMock := mock_actor.NewMockServerGateway(mockCtrl)
	storage := actor.NewMemoryStorage()

	config := actor.Config{
		Servers:                    []actor.ServerID{localServerID},
		LeaderElectionTimeout:      10,
		LeaderElectionTimeoutSplay: 0,
		LeaderHeartbeatFrequency:   5,
	}

	server := NewServer(localServerID, storage, gatewayMock, config)

	startReturned := false
	stopReturned := false

	go func() {
		server.Start()
		startReturned = true
	}()

	go func() {
		server.Stop()
		stopReturned = true
	}()

	for {
		if startReturned && stopReturned {
			return
		}
	}
}
