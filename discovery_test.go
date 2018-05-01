package raft_test

import (
	"github.com/sema/raft"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStaticDiscovery_Quorum(t *testing.T) {
	discovery := raft.NewStaticDiscovery([]raft.ServerID{"server1"})
	assert.Equal(t, 1, discovery.Quorum())

	discovery = raft.NewStaticDiscovery([]raft.ServerID{"server1", "server2"})
	assert.Equal(t, 2, discovery.Quorum())

	discovery = raft.NewStaticDiscovery([]raft.ServerID{"server1", "server2", "server3"})
	assert.Equal(t, 2, discovery.Quorum())

	discovery = raft.NewStaticDiscovery([]raft.ServerID{"server1", "server2", "server3", "server4"})
	assert.Equal(t, 3, discovery.Quorum())
}
