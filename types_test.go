package go_raft_test

import (
	"github.com/sema/go-raft"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMaxLogIndex(t *testing.T) {
	assert.Equal(t, go_raft.LogIndex(3), go_raft.MaxLogIndex(2, 3))
	assert.Equal(t, go_raft.LogIndex(3), go_raft.MaxLogIndex(3, 2))
}

func TestMinLogIndex(t *testing.T) {
	assert.Equal(t, go_raft.LogIndex(2), go_raft.MinLogIndex(2, 3))
	assert.Equal(t, go_raft.LogIndex(2), go_raft.MinLogIndex(3, 2))
}
