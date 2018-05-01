package integration_test

import (
	"github.com/sema/raft"
	"testing"
)

const server1ID = raft.ServerID("server1.raft.local")
const server2ID = raft.ServerID("server2.raft.local")
const server3ID = raft.ServerID("server3.raft.local")

func TestIntegration__IsAbleToElectLeader(t *testing.T) {
	serverIDs := []raft.ServerID{server1ID, server2ID, server3ID}

	config := raft.Config{
		LeaderElectionTimeout:      4,
		LeaderElectionTimeoutSplay: 4,
		LeaderHeartbeatFrequency:   2,
	}

	servers := make(map[raft.ServerID]raft.Server)

	for _, serverID := range serverIDs {
		gateway := raft.NewLocalServerGateway(servers)
		discovery := raft.NewStaticDiscovery(serverIDs)
		persistentStorage := raft.NewMemoryStorage()

		server := raft.NewServer(serverID, persistentStorage, gateway, discovery, config)

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
