package go_raft_test

import (
	"testing"
	"github.com/sema/go-raft"
	"github.com/stretchr/testify/assert"
)

func TestMemoryStorage_MergeLogs_RemovesEntriesAfterInputIfTermsDiffer(t *testing.T) {
	storage := go_raft.NewMemoryStorage()

	storage.SetCurrentTerm(0)
	storage.AppendLog("")  // index 1, term 0
	storage.AppendLog("")  // index 2, term 0
	storage.AppendLog("")  // index 3, term 0

	storage.MergeLogs([]go_raft.LogEntry{
		go_raft.NewLogEntry(1, 2, ""),
	})

	logEntry := storage.LatestLogEntry()
	assert.Equal(t, go_raft.LogIndex(2), logEntry.Index)
	assert.Equal(t, go_raft.Term(1), logEntry.Term)
	assert.Equal(t, 2, storage.LogLength())
}
