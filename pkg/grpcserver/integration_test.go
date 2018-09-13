package grpcserver

import (
	"testing"

	"fmt"
	"io/ioutil"

	"os"

	"github.com/sema/raft/pkg/actor"
	"github.com/sema/raft/pkg/server"
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

	servers, cancelFunc := testSetup3Servers(t)
	defer cancelFunc()

	testWaitUntilLeaderIsElected(t, servers)
}

func TestIntegration__ProposedChangesArePropagatedToAllServers(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	servers, cleanupFunc := testSetup3Servers(t)
	defer cleanupFunc()

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

	servers, cleanupFunc := testSetup3Servers(t)
	defer cleanupFunc()

	oldLeader := testWaitUntilLeaderIsElected(t, servers)

	servers[oldLeader].Stop()

	for {
		if oldLeader != testWaitUntilLeaderIsElected(t, servers) {
			return
		}
	}
}

func testSetup3Servers(t *testing.T) (map[actor.ServerID]server.Server, func()) {
	config := actor.Config{
		Servers:                    []actor.ServerID{server1ID, server2ID, server3ID},
		LeaderElectionTimeout:      4,
		LeaderElectionTimeoutSplay: 4,
		LeaderHeartbeatFrequency:   2,
	}

	tmpDir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)

	discovery := make(map[actor.ServerID]DiscoveryConfig)
	for _, serverID := range config.Servers {
		socketPath := fmt.Sprintf("unix:%s/%s.socket", tmpDir, serverID)
		discovery[serverID] = DiscoveryConfig{
			ServerID:      actor.ServerID(serverID),
			AddressRemote: socketPath,
		}
	}

	servers := make(map[actor.ServerID]server.Server)
	serversOutbound := make(map[actor.ServerID]*OutboundGRPCServer)
	serversInbound := make(map[actor.ServerID]*InboundGRPCServer)

	for _, serverID := range config.Servers {
		storage := actor.NewMemoryStorage()

		svr := server.NewServer(serverID, storage, config)
		go svr.Start()

		inboundServer := NewInboundGRPCServer(svr)
		go inboundServer.Serve(discovery[serverID].AddressRemote)

		outboundServer := NewOutboundGRPCServer(svr, discovery)
		go outboundServer.Serve()

		servers[serverID] = svr
		serversOutbound[serverID] = outboundServer
		serversInbound[serverID] = inboundServer
	}

	cleanupFunc := func() {
		for _, serverID := range config.Servers {
			servers[serverID].Stop()
			serversOutbound[serverID].Stop()
			serversInbound[serverID].Stop()
		}
		os.RemoveAll(tmpDir)
	}

	return servers, cleanupFunc
}

func testWaitUntilLeaderIsElected(t *testing.T, servers map[actor.ServerID]server.Server) actor.ServerID {
	for {
		for serverID, svr := range servers {
			if svr.CurrentStateName() == "LeaderMode" {
				return serverID
			}

			testCheckTickLimit(t, svr)
		}
	}
}

func testCheckTickLimit(t *testing.T, svr server.Server) {
	if svr.Age() > tickLimit {
		t.Errorf("Test timed out as server %s exceeded tick limit of %d", svr, tickLimit)
	}
}

func testWaitForLogEntryToReplicate(t *testing.T, servers map[actor.ServerID]server.Server, payload string) {
	for {
		logEntryOnAllServers := true
		for _, svr := range servers {
			latestCommitEntry, ok := svr.Log(svr.CommitIndex())
			assert.True(t, ok)

			if latestCommitEntry.Payload != payload {
				logEntryOnAllServers = false
			}

			testCheckTickLimit(t, svr)
		}

		if logEntryOnAllServers {
			return
		}
	}
}
