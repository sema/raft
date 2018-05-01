package raft_test

import (
	"github.com/sema/raft"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMemoryStorage_MergeLogs_RemovesEntriesAfterInputIfTermsDiffer(t *testing.T) {
	storage := raft.NewMemoryStorage()

	storage.SetCurrentTerm(0)
	storage.AppendLog("") // index 1, term 0
	storage.AppendLog("") // index 2, term 0
	storage.AppendLog("") // index 3, term 0

	storage.MergeLogs([]raft.LogEntry{
		raft.NewLogEntry(1, 2, ""),
	})

	logEntry := storage.LatestLogEntry()
	assert.Equal(t, raft.LogIndex(2), logEntry.Index)
	assert.Equal(t, raft.Term(1), logEntry.Term)
	assert.Equal(t, 2, storage.LogLength())
}
