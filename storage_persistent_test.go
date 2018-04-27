package go_raft_test

import (
	"testing"
	"github.com/sema/go-raft"
	"github.com/stretchr/testify/assert"
)

func TestMemoryStorage_PruneLogEntriesAfter_RemovesAllEntriesAfterGivenIndex(t *testing.T) {
	storage := go_raft.NewMemoryStorage()

	storage.AppendLog("")  // index 1
	storage.AppendLog("")  // index 2
	storage.AppendLog("")  // index 3

	storage.PruneLogEntriesAfter(go_raft.LogIndex(2))

	logEntry := storage.LatestLogEntry()
	assert.Equal(t, go_raft.LogIndex(2), logEntry.Index)
}

func TestMemoryStorage_PruneLogEntriesAfter_RemovesAllEntriesWhenGivenIndex0(t *testing.T) {
	storage := go_raft.NewMemoryStorage()

	storage.AppendLog("")  // index 1
	storage.AppendLog("")  // index 2
	storage.AppendLog("")  // index 3

	storage.PruneLogEntriesAfter(go_raft.LogIndex(0))

	logEntry := storage.LatestLogEntry()
	assert.Equal(t, go_raft.LogIndex(0), logEntry.Index)
}
