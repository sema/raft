package actor

import (
	"testing"

	"github.com/sema/raft"
	"github.com/stretchr/testify/assert"
)

func TestMaxLogIndex(t *testing.T) {
	assert.Equal(t, raft.LogIndex(3), raft.MaxLogIndex(2, 3))
	assert.Equal(t, raft.LogIndex(3), raft.MaxLogIndex(3, 2))
}

func TestMinLogIndex(t *testing.T) {
	assert.Equal(t, raft.LogIndex(2), raft.MinLogIndex(2, 3))
	assert.Equal(t, raft.LogIndex(2), raft.MinLogIndex(3, 2))
}
