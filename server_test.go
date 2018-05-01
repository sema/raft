package raft_test

import (
	"testing"
	"github.com/sema/raft/mocks"
	"github.com/golang/mock/gomock"
	"github.com/sema/raft"
)

func TestServer_StopCausesStartToReturn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	gatewayMock := mock_raft.NewMockServerGateway(mockCtrl)
	storage := raft.NewMemoryStorage()

	config := raft.Config{
		Servers:                    []raft.ServerID{localServerID},
		LeaderElectionTimeout:      10,
		LeaderElectionTimeoutSplay: 0,
		LeaderHeartbeatFrequency:   5,
	}

	server := raft.NewServer(localServerID, storage, gatewayMock, config)

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