package raft_test

import (
	"github.com/sema/raft"
	"testing"
)

const server1ID = raft.ServerID("server1.raft.local")
const server2ID = raft.ServerID("server2.raft.local")
const server3ID = raft.ServerID("server3.raft.local")

type localGateway struct {
	servers map[raft.ServerID]raft.Server
}

func newLocalServerGateway() *localGateway {
	return &localGateway{
		servers: map[raft.ServerID]raft.Server{},
	}
}

func (g *localGateway) RegisterServer(serverID raft.ServerID, server raft.Server) {
	g.servers[serverID] = server
}

func (g *localGateway) Send(to raft.ServerID, message raft.Message) {
	server := g.servers[to]
	server.SendMessage(message)
}

func TestIntegration__IsAbleToElectLeader(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	config := raft.Config{
		Servers:                    []raft.ServerID{server1ID, server2ID, server3ID},
		LeaderElectionTimeout:      4,
		LeaderElectionTimeoutSplay: 4,
		LeaderHeartbeatFrequency:   2,
	}

	gateway := newLocalServerGateway()

	servers := make(map[raft.ServerID]raft.Server)
	for _, serverID := range config.Servers {
		storage := raft.NewMemoryStorage()
		server := raft.NewServer(serverID, storage, gateway, config)
		gateway.RegisterServer(serverID, server)

		servers[serverID] = server
	}

	for _, server := range servers {
		go server.Run()
	}

	for {
		for _, server := range servers {
			if server.CurrentStateName() == "LeaderMode" {
				return
			}

		}
	}
}

// TODO Write integration tests
// - proposed changes are propagated to all instances and committed
// - new leader is elected if a node dies
