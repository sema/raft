package raft

import (
	"log"
)

const NoVote ServerID = ""

// Storage defines the interface for persistent storage needed by the Raft algorithm. An implementation of
// Storage should ensure any values written are persisted before returning. While all write access to the storage
// happens from the same thread, read operations may occur from multiple threads concurrently.
//
// The API is purposely simplified to make it easier to implement. For example, invariants between term and votedFor
// are handled externally by callers to reduce logic in individual implementations.
type Storage interface {
	CurrentTerm() Term
	SetCurrentTerm(newTerm Term)

	VotedFor() ServerID
	SetVotedFor(votedFor ServerID)
	UnsetVotedFor()

	Log(index LogIndex) (logEntry LogEntry, ok bool)
	LogRange(startIndex LogIndex) []LogEntry
	LatestLogEntry() (logEntry LogEntry)

	AppendLog(payload string)
	MergeLogs([]LogEntry)
}

type memoryStorage struct {
	currentTerm Term
	votedFor    ServerID
	logEntries  []LogEntry
}

// NewMemoryStorage returns a memory backed implementation of the Storage interface for testing purposes. This
// implementation does not persist values, making it unusable for real world use cases.
func NewMemoryStorage() Storage {
	return &memoryStorage{
		votedFor: NoVote,
	}
}

// CurrentTerm returns the current term
func (ms *memoryStorage) CurrentTerm() Term {
	return ms.currentTerm
}

// SetCurrentTerm sets the current term.
func (ms *memoryStorage) SetCurrentTerm(newTerm Term) {
	if newTerm < ms.currentTerm {
		log.Panicf("CurrentTerm is not allowed to decrease (%d -> %d).", ms.currentTerm, newTerm)
	}

	ms.currentTerm = newTerm
}

// VoteFor returns the server voted for as leader in the current term.
func (ms *memoryStorage) VotedFor() ServerID {
	return ms.votedFor
}

// SetVotedFor assigns a vote to a server within the current term. This method panics if called more than once within a
// term (i.e. when a vote has already been cast).
func (ms *memoryStorage) SetVotedFor(votedFor ServerID) {
	if ms.votedFor != NoVote {
		log.Panicf("Overwriting VoteFor value is not allowed, value already set to %s", ms.votedFor)
	}

	ms.votedFor = votedFor
}

// UnsetVotedFor removes the current vote. This is called when incrementing the term.
func (ms *memoryStorage) UnsetVotedFor() {
	ms.votedFor = NoVote
}

// Log returns the log entry at a given index.
func (ms *memoryStorage) Log(index LogIndex) (entry LogEntry, ok bool) {
	if index == 0 {
		// Base case - no previous log entries in log
		return NewLogEntry(0, 0, ""), true
	}

	index = index - 1 // convert from 1 indexed to 0 indexed

	if int(index) >= len(ms.logEntries) {
		return LogEntry{}, false
	}

	return ms.logEntries[index], true
}

// AppendLog appends a new log entry with the current term and next index.
func (ms *memoryStorage) AppendLog(payload string) {
	entry := LogEntry{
		Term:    ms.currentTerm,
		Index:   ms.LatestLogEntry().Index + 1,
		Payload: payload,
	}

	ms.logEntries = append(ms.logEntries, entry)
}

// LatestLogEntry returns the latest log entry, or the sentinel log entry if none exist. The sentinel log entry
// has the term 0 and index 0.
func (ms *memoryStorage) LatestLogEntry() (logEntry LogEntry) {
	if len(ms.logEntries) == 0 {
		return LogEntry{
			Term:  0,
			Index: 0,
		}
	}

	return ms.logEntries[len(ms.logEntries)-1]
}

func (ms *memoryStorage) pruneLogEntriesAfter(index LogIndex) {
	index = index - 1 // convert from 1 indexed to 0 indexed
	ms.logEntries = ms.logEntries[:index+1]
}

// MergeLogs merges in an input sequence of log entries into the current log.
//
// Entries are merged in by their index, and may append to the log if
// new entries are provided in the input.
//
// If two entries with identical indexes differ by their term, then the input
// entry overwrites the current log entry, and all subsequent log entries with
// a higher index are pruned from the current log. Any additional input log
// entries are then appended to the current log.
//
// NOTE: The raft algorithm guarantees that if two entries have the same
// index and term, then the payload is also identical. This guarantee also
// applies to entries in the current and input log with identical term and
// index.
//
// The provided entries MUST have sequential indexes, and there MUST NOT
// be a gap in indexes between the last entry of the current log and the
// first entry of the new input.
//
// Example valid indexes:
// Existing: 1, 2, 3 - Input: 4, 5, 6
// Existing: 1, 2, 3 - Input: 2, 3
//
// Example invalid indexes:
// Existing 1, 2, 3 - Input: 1, 3, 4 (gap between 1 and 3)
// Existing 1, 2, 3 - Input: 5, 6 (gap between 3 and 5)
//
func (ms *memoryStorage) MergeLogs(entries []LogEntry) {
	lastIndex := ms.LatestLogEntry().Index

	for _, entry := range entries {
		if entry.Index <= lastIndex {
			// Merge existing entry

			// Convert from 1 indexed to 0 indexed
			zeroedIndex := entry.Index - 1

			if ms.logEntries[zeroedIndex].Term != entry.Term {
				// Only merge if terms differ, otherwise Raft guarantees that this entry
				// and all following entries match
				ms.pruneLogEntriesAfter(entry.Index)
				ms.logEntries[zeroedIndex] = entry
				lastIndex = entry.Index
			}

		} else if entry.Index == lastIndex+1 {
			// Append
			ms.logEntries = append(ms.logEntries, entry)
			lastIndex++

		} else {
			log.Panicf("Trying to append log entries starting with index %d to logs ending "+
				"with index %d", entry.Index, lastIndex-1)
		}
	}
}

func (ms *memoryStorage) LogRange(startIndex LogIndex) []LogEntry {
	if startIndex > ms.LatestLogEntry().Index {
		return nil
	}
	return ms.logEntries[int(startIndex)-1:]
}
