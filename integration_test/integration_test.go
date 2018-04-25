package integration_test

import (
	"testing"
	"github/sema/go-raft"
	"time"
	"log"
	"os"
)

const server1ID = go_raft.ServerID("server1.servers.local")
const server2ID = go_raft.ServerID("server2.servers.local")


func TestIntegration(t *testing.T) {
	log.SetOutput(os.Stdout)

	serverIDs := []go_raft.ServerID{server1ID, server2ID}

	config := go_raft.Config{
		LeaderElectionTimeout: 100 * time.Millisecond,
		LeaderElectionTimeoutSplay: 10 * time.Millisecond,
	}

	servers := make([]go_raft.Server, 0)
	serversm := make(map[go_raft.ServerID]go_raft.Server)

	for _, serverID := range serverIDs {
		gateway := go_raft.NewLocalServerGateway(serversm)
		discovery := go_raft.NewStaticDiscovery(serverIDs)
		persistentStorage := go_raft.NewMemoryStorage()

		server := go_raft.NewServer(serverID, persistentStorage, gateway, discovery, config)

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

	// TODO Wait until one of the serverIDs is elected as leader, and exit
	// TODO change all logic to not be timing but tick based, such that we can provide a non-time
	// based bound on the test time
}
