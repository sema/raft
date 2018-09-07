package raft

import (
	"testing"

	"github.com/sema/raft/pkg/actor"
	"github.com/stretchr/testify/assert"
)

const server1ID = actor.ServerID("server1.raft.local")
const server2ID = actor.ServerID("server2.raft.local")
const server3ID = actor.ServerID("server3.raft.local")
const tickLimit = actor.Tick(500)

func TestIntegration__IsAbleToElectLeader(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	servers, _ := testSetup3Servers(t)
	testWaitUntilLeaderIsElected(t, servers)
}

func TestIntegration__ProposedChangesArePropagatedToAllServers(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	servers, _ := testSetup3Servers(t)
	leader := testWaitUntilLeaderIsElected(t, servers)

	leaderServer := servers[leader]
	err := leaderServer.SendMessage(actor.NewMessageProposal(leader, leader, "newLogEntry"))
	assert.NoError(t, err)

	testWaitForLogEntryToReplicate(t, servers, "newLogEntry")
}

func TestIntegration__NewLeaderIsElectedIfNodeFails(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	servers, gateway := testSetup3Servers(t)
	oldLeader := testWaitUntilLeaderIsElected(t, servers)

	gateway.ChangeServerConnectivity(oldLeader, false)

	for {
		if oldLeader != testWaitUntilLeaderIsElected(t, servers) {
			return
		}
	}
}

func testSetup3Servers(t *testing.T) (map[actor.ServerID]Server, *localGateway) {
	config := actor.Config{
		Servers:                    []actor.ServerID{server1ID, server2ID, server3ID},
		LeaderElectionTimeout:      4,
		LeaderElectionTimeoutSplay: 4,
		LeaderHeartbeatFrequency:   2,
	}

	gateway := newLocalServerGateway(t)

	servers := make(map[actor.ServerID]Server)
	for _, serverID := range config.Servers {
		storage := actor.NewMemoryStorage()
		server := NewServer(serverID, storage, gateway, config)
		gateway.RegisterServer(serverID, server)

		servers[serverID] = server
	}

	for _, server := range servers {
		go server.Start()
	}

	return servers, gateway
}

func testWaitUntilLeaderIsElected(t *testing.T, servers map[actor.ServerID]Server) actor.ServerID {
	for {
		for serverID, server := range servers {
			if server.CurrentStateName() == "LeaderMode" {
				return serverID
			}

			testCheckTickLimit(t, server)
		}
	}
}

func testCheckTickLimit(t *testing.T, server Server) {
	if server.Age() > tickLimit {
		t.Errorf("Test timed out as server %s exceeded tick limit of %d", server, tickLimit)
	}
}

func testWaitForLogEntryToReplicate(t *testing.T, servers map[actor.ServerID]Server, payload string) {
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
	servers       map[actor.ServerID]Server
	downedServers map[actor.ServerID]bool
	t             *testing.T
}

func newLocalServerGateway(t *testing.T) *localGateway {
	return &localGateway{
		servers:       map[actor.ServerID]Server{},
		downedServers: map[actor.ServerID]bool{},
		t:             t,
	}
}

func (g *localGateway) RegisterServer(serverID actor.ServerID, server Server) {
	g.servers[serverID] = server
}

func (g *localGateway) ChangeServerConnectivity(serverID actor.ServerID, connected bool) {
	g.downedServers[serverID] = !connected
}

func (g *localGateway) Send(to actor.ServerID, message actor.Message) {
	if g.downedServers[to] {
		return // drop all traffic to downed server
	}

	if g.downedServers[message.From] {
		return // drop all traffic from downed server
	}

	server := g.servers[to]
	err := server.SendMessage(message)
	assert.NoError(g.t, err)
}
