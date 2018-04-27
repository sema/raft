package go_raft

import (
	"log"
)

// TODO rename this back to storage and change volatile to something else?

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
	// TODO constant? ok response?
	if ms.votedFor == "" {
		ms.votedFor = votedFor
	}
}

func (ms *memoryStorage) ClearVotedFor() {
	ms.votedFor = ""
}

func (ms *memoryStorage) Log(index LogIndex) (LogEntry, bool) {
	index = index - 1 // convert from 1 indexed to 0 indexed

	if index < 0 {
		return LogEntry{  // special sentinel value
			Term: 0,
			Index: 0,
		}, true
	}

	if int(index) >= len(ms.logEntries) {
		return LogEntry{}, false
	}

	return ms.logEntries[index-1], true
}

func (ms *memoryStorage) AppendLog(payload string) {
	entry := LogEntry{
		Term:  ms.currentTerm,
		Index: LogIndex(len(ms.logEntries) + 1),
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

func (ms *memoryStorage) PruneLogEntriesAfter(index LogIndex) {
	index = index - 1 // convert from 1 indexed to 0 indexed

	ms.logEntries = ms.logEntries[:index+1]
}

func (ms *memoryStorage) AppendLogs(entries []LogEntry) {
	if len(entries) > 0 {
		lastOldIndex := ms.LatestLogEntry().Index
		firstNewIndex := entries[0].Index
		if lastOldIndex +1 != firstNewIndex {
			log.Panicf("Trying to append log entries starting with index %d to logs ending with index %d", firstNewIndex, lastOldIndex)
		}
	}

	for _, entry := range entries {
		ms.logEntries = append(ms.logEntries, entry)
	}
}