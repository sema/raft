package raft_test

import (
	"github.com/sema/raft"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig_Quorum(t *testing.T) {
	config := raft.Config{
		Servers: []raft.ServerID{"server1"},
	}
	assert.Equal(t, 1, config.Quorum())

	config = raft.Config{
		Servers: []raft.ServerID{"server1", "server2"},
	}
	assert.Equal(t, 2, config.Quorum())

	config = raft.Config{
		Servers: []raft.ServerID{"server1", "server2", "server3"},
	}
	assert.Equal(t, 2, config.Quorum())

	config = raft.Config{
		Servers: []raft.ServerID{"server1", "server2", "server3", "server4"},
	}
	assert.Equal(t, 3, config.Quorum())
}
