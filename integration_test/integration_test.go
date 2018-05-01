package integration_test

import (
	"github/sema/raft"
	"log"
	"os"
	"testing"
	"time"
)

const server1ID = raft.ServerID("server1.servers.local")
const server2ID = raft.ServerID("server2.servers.local")

func TestIntegration(t *testing.T) {
	log.SetOutput(os.Stdout)

	serverIDs := []raft.ServerID{server1ID, server2ID}

	config := raft.Config{
		LeaderElectionTimeout:      100 * time.Millisecond,
		LeaderElectionTimeoutSplay: 10 * time.Millisecond,
	}

	servers := make([]raft.Server, 0)
	serversm := make(map[raft.ServerID]raft.Server)

	for _, serverID := range serverIDs {
		gateway := raft.NewLocalServerGateway(serversm)
		discovery := raft.NewStaticDiscovery(serverIDs)
		persistentStorage := raft.NewMemoryStorage()

		server := raft.NewServer(serverID, persistentStorage, gateway, discovery, config)

		servers = append(servers, server)
		serversm[serverID] = server
	}

	for _, server := range servers {
		go server.Run()
	}

	for {
		for _, server := range servers {
			if server.CurrentStateName() == "leader2" {
				return
			}
		}
	}

	// TODO Write integration tests
}
