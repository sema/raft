package actor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryStorage_MergeLogs_OverwritesEntriesIfTermsDiffer(t *testing.T) {
	storage := testStorageWith3EntriesInTerm0()

	// index 1, term 0
	// index 2, term 0
	// index 3, term 0

	storage.MergeLogs([]LogEntry{
		NewLogEntry(1, 2, ""),
		NewLogEntry(1, 3, ""),
	})

	logEntry := storage.LatestLogEntry()
	assert.Equal(t, LogIndex(3), logEntry.Index)
	assert.Equal(t, Term(1), logEntry.Term)
}

func TestMemoryStorage_MergeLogs_AppendsNewEntries(t *testing.T) {
	storage := testStorageWith3EntriesInTerm0()

	// index 1, term 0
	// index 2, term 0
	// index 3, term 0

	storage.MergeLogs([]LogEntry{
		NewLogEntry(1, 4, ""),
		NewLogEntry(1, 5, ""),
	})

	logEntry := storage.LatestLogEntry()
	assert.Equal(t, LogIndex(5), logEntry.Index)
	assert.Equal(t, Term(1), logEntry.Term)
}

func TestMemoryStorage_MergeLogs_PrunesAnyExistingEntriesAfterATermMismatch(t *testing.T) {
	storage := testStorageWith3EntriesInTerm0()

	// index 1, term 0
	// index 2, term 0
	// index 3, term 0

	storage.MergeLogs([]LogEntry{
		NewLogEntry(1, 1, ""),
		NewLogEntry(1, 2, ""),
	})

	logEntry := storage.LatestLogEntry()
	assert.Equal(t, LogIndex(2), logEntry.Index)
	assert.Equal(t, Term(1), logEntry.Term)
}

func TestMemoryStorage_MergeLogs_TryingToMergeLogsWithGapsInIndexPanics(t *testing.T) {
	storage := testStorageWith3EntriesInTerm0()

	assert.Panics(t, func() {
		storage.MergeLogs([]LogEntry{
			NewLogEntry(1, 10, ""),
		})
	})

	assert.Panics(t, func() {
		storage.MergeLogs([]LogEntry{
			NewLogEntry(0, 1, ""),
			NewLogEntry(1, 10, ""),
		})
	})
}

func TestMemoryStorage_LogReturnsFalseIfIndexIsOutOfRange(t *testing.T) {
	storage := testStorageWith3EntriesInTerm0()
	_, ok := storage.Log(LogIndex(4))
	assert.False(t, ok)
}

func TestMemoryStorage_LogOfZeroIndexReturnsSentinelValue(t *testing.T) {
	storage := testStorageWith3EntriesInTerm0()
	entry, ok := storage.Log(LogIndex(0))

	assert.True(t, ok)
	assert.Equal(t, LogIndex(0), entry.Index)
	assert.Equal(t, Term(0), entry.Term)
}

func TestMemoryStorage_LogReturnsExpectedValue(t *testing.T) {
	storage := testStorageWith3EntriesInTerm0()
	entry, ok := storage.Log(LogIndex(1))

	assert.True(t, ok)
	assert.Equal(t, LogIndex(1), entry.Index)
	assert.Equal(t, Term(0), entry.Term)
}

func TestMemoryStorage_AppendLogAppendsLogEntryWithCurrentTermAndIncrementedIndex(t *testing.T) {
	storage := testStorageWith3EntriesInTerm0()
	storage.AppendLog("something")

	latestEntry := storage.LatestLogEntry()
	assert.Equal(t, Term(0), latestEntry.Term)
	assert.Equal(t, LogIndex(4), latestEntry.Index)
}

func TestMemoryStorage_LatestLogEntryReturnsLatestEntry(t *testing.T) {
	storage := testStorageWith3EntriesInTerm0()

	latestEntry := storage.LatestLogEntry()
	assert.Equal(t, Term(0), latestEntry.Term)
	assert.Equal(t, LogIndex(3), latestEntry.Index)
}

func TestMemoryStorage_LogRangeReturnsARangeOfEntries(t *testing.T) {
	storage := testStorageWith3EntriesInTerm0()
	logEntries := storage.LogRange(LogIndex(2))

	assert.Equal(t, []LogEntry{
		NewLogEntry(0, 2, ""),
		NewLogEntry(0, 3, ""),
	}, logEntries)
}

func testStorageWith3EntriesInTerm0() Storage {
	storage := NewMemoryStorage()

	storage.SetCurrentTerm(0)
	storage.AppendLog("") // index 1, term 0
	storage.AppendLog("") // index 2, term 0
	storage.AppendLog("") // index 3, term 0

	return storage
}
