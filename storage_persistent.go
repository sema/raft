package go_raft

import (
	"log"
)

// TODO go through storage and cleanup API
// TODO test for proper handling of edge cases in storage

const NO_VOTE ServerID = ""

// memoryStorage implements the PersistentStorage interface using a memory back persistentStorage. Should only be used for testing!
type memoryStorage struct {
	currentTerm Term
	votedFor    ServerID
	logEntries  []LogEntry
}

func NewMemoryStorage() PersistentStorage {
	return &memoryStorage{}
}

func (ms *memoryStorage) CurrentTerm() Term {
	return ms.currentTerm
}

func (ms *memoryStorage) SetCurrentTerm(newTerm Term) {
	if newTerm > ms.currentTerm {
		// Don't clear vote if we are setting the same Term multiple times, which might occur in certain edge cases
		ms.votedFor = ""
	}
	ms.currentTerm = newTerm
}

func (ms *memoryStorage) VotedFor() ServerID {
	return ms.votedFor
}

func (ms *memoryStorage) SetVotedForIfUnset(votedFor ServerID) {
	if ms.votedFor == NO_VOTE {
		ms.votedFor = votedFor
	}
}

func (ms *memoryStorage) ClearVotedFor() {
	ms.votedFor = NO_VOTE
}

func (ms *memoryStorage) Log(index LogIndex) (LogEntry, bool) {
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

func (ms *memoryStorage) AppendLog(payload string) {
	entry := LogEntry{
		Term:    ms.currentTerm,
		Index:   LogIndex(len(ms.logEntries) + 1),
		Payload: payload,
	}

	ms.logEntries = append(ms.logEntries, entry)
}

func (ms *memoryStorage) LatestLogEntry() (logEntry LogEntry) {
	if len(ms.logEntries) == 0 {
		return LogEntry{
			Term:  0,
			Index: 0,
		}
	}

	return ms.logEntries[len(ms.logEntries)-1]
}

func (ms *memoryStorage) LogLength() int {
	return len(ms.logEntries)
}

func (ms *memoryStorage) pruneLogEntriesAfter(index LogIndex) {
	index = index - 1 // convert from 1 indexed to 0 indexed
	ms.logEntries = ms.logEntries[:index+1]
}

func (ms *memoryStorage) MergeLogs(entries []LogEntry) {
	nextIndex := ms.LatestLogEntry().Index + 1

	for _, entry := range entries {
		if entry.Index < nextIndex {
			// Merge existing entry
			index := entry.Index - 1
			if ms.logEntries[index].Term != entry.Term {
				// Only merge if terms differ, otherwise Raft guarantees that this entry and all following entries match
				ms.pruneLogEntriesAfter(entry.Index)
				ms.logEntries[index] = entry
			}
		} else if entry.Index == nextIndex {
			// Append
			ms.logEntries = append(ms.logEntries, entry)
			nextIndex += 1
		} else {
			log.Panicf("Trying to append log entries starting with index %d to logs ending with index %d", entry.Index, nextIndex-1)
		}
	}
}

func (ms *memoryStorage) LogRange(startIndex LogIndex) []LogEntry {
	startIndex = startIndex - 1 // zero index

	if int(startIndex) >= len(ms.logEntries) {
		return nil
	}
	return ms.logEntries[startIndex:]
}
