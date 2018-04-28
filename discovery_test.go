package go_raft_test

import (
	"testing"
	"github.com/sema/go-raft"
	"github.com/stretchr/testify/assert"
)

func TestStaticDiscovery_Quorum(t *testing.T) {
	var discovery go_raft.ServerDiscovery

	discovery = go_raft.NewStaticDiscovery([]go_raft.ServerID{"server1"})
	assert.Equal(t, 1, discovery.Quorum())

	discovery = go_raft.NewStaticDiscovery([]go_raft.ServerID{"server1", "server2"})
	assert.Equal(t, 2, discovery.Quorum())

	discovery = go_raft.NewStaticDiscovery([]go_raft.ServerID{"server1", "server2", "server3"})
	assert.Equal(t, 2, discovery.Quorum())

	discovery = go_raft.NewStaticDiscovery([]go_raft.ServerID{"server1", "server2", "server3", "server4"})
	assert.Equal(t, 3, discovery.Quorum())
}