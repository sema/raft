package raft_test

import (
	"github.com/sema/raft"
	"testing"
	"github.com/stretchr/testify/assert"
)

const server1ID = raft.ServerID("server1.raft.local")
const server2ID = raft.ServerID("server2.raft.local")
const server3ID = raft.ServerID("server3.raft.local")
const tickLimit = raft.Tick(500)

func TestIntegration__IsAbleToElectLeader(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	servers, _ := testSetup3Servers()
	testWaitUntilLeaderIsElected(t, servers)
}

func TestIntegration__ProposedChangesArePropagatedToAllServers(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	servers, _ := testSetup3Servers()
	leader := testWaitUntilLeaderIsElected(t, servers)

	leaderServer := servers[leader]
	leaderServer.SendMessage(raft.NewMessageProposal(leader, leader, "newLogEntry"))

	testWaitForLogEntryToReplicate(t, servers, "newLogEntry")
}

func TestIntegration__NewLeaderIsElectedIfNodeFails(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	servers, gateway := testSetup3Servers()
	oldLeader := testWaitUntilLeaderIsElected(t, servers)

	gateway.ChangeServerConnectivity(oldLeader, false)

	for {
		if oldLeader != testWaitUntilLeaderIsElected(t, servers) {
			return
		}
	}
}

func testSetup3Servers() (map[raft.ServerID]raft.Server, *localGateway) {
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

	return servers, gateway
}

func testWaitUntilLeaderIsElected(t *testing.T, servers map[raft.ServerID]raft.Server) raft.ServerID {
	for {
		for serverID, server := range servers {
			if server.CurrentStateName() == "LeaderMode" {
				return serverID
			}

			testCheckTickLimit(t, server)
		}
	}
}

func testCheckTickLimit(t *testing.T, server raft.Server) {
	if server.Age() > tickLimit {
		t.Errorf("Test timed out as server %s exceeded tick limit of %d", server, tickLimit)
	}
}

func testWaitForLogEntryToReplicate(t *testing.T, servers map[raft.ServerID]raft.Server, payload string) {
	for {
		logEntryOnAllServers := true
		for _, server := range servers {
			latestCommitEntry, ok := server.Log(server.CommitIndex())
			assert.True(t, ok)

			if latestCommitEntry.Payload != payload {
				logEntryOnAllServers = false
			}

			testCheckTickLimit(t, server)
		}

		if logEntryOnAllServers {
			return
		}
	}
}

type localGateway struct {
	servers map[raft.ServerID]raft.Server
	downedServers map[raft.ServerID]bool
}

func newLocalServerGateway() *localGateway {
	return &localGateway{
		servers: map[raft.ServerID]raft.Server{},
		downedServers: map[raft.ServerID]bool{},
	}
}

func (g *localGateway) RegisterServer(serverID raft.ServerID, server raft.Server) {
	g.servers[serverID] = server
}

func (g *localGateway) ChangeServerConnectivity(serverID raft.ServerID, connected bool) {
	g.downedServers[serverID] = !connected
}

func (g *localGateway) Send(to raft.ServerID, message raft.Message) {
	if g.downedServers[to] {
		return  // drop all traffic to downed server
	}

	if g.downedServers[message.From] {
		return  // drop all traffic from downed server
	}

	server := g.servers[to]
	server.SendMessage(message)
}
