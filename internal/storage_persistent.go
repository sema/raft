package internal

import "github/sema/go-raft"

// TODO rename this back to storage and change volatile to something else?

// memoryStorage implements the PersistentStorage interface using a memory back persistentStorage. Should only be used for testing!
type memoryStorage struct {
	currentTerm go_raft.Term
	votedFor    go_raft.ServerID
	logEntries  []go_raft.LogEntry
}

func NewMemoryStorage() go_raft.PersistentStorage {
	return &memoryStorage{}
}

func (ms *memoryStorage) CurrentTerm() go_raft.Term {
	return ms.currentTerm
}

func (ms *memoryStorage) SetCurrentTerm(newTerm go_raft.Term) {
	if newTerm > ms.currentTerm {
		// Don't clear vote if we are setting the same term multiple times, which might occur in certain edge cases
		ms.votedFor = ""
	}
	ms.currentTerm = newTerm
}

func (ms *memoryStorage) VotedFor() go_raft.ServerID {
	return ms.votedFor
}

func (ms *memoryStorage) SetVotedForIfUnset(votedFor go_raft.ServerID) {
	// TODO constant? ok response?
	if ms.votedFor == "" {
		ms.votedFor = votedFor
	}
}

func (ms *memoryStorage) ClearVotedFor() {
	ms.votedFor = ""
}

func (ms *memoryStorage) Log(index go_raft.LogIndex) (go_raft.LogEntry, bool) {
	// TODO this need to be atomic?
	if index == 0 {
		return go_raft.LogEntry{}, false
	}

	if len(ms.logEntries) < int(index) {
		return go_raft.LogEntry{}, false
	}

	return ms.logEntries[index-1], true
}

func (ms *memoryStorage) AppendLog(entry LogEntry) {
	// TODO index check?
	ms.logEntries = append(ms.logEntries, entry)
}

func (ms *memoryStorage) MergeLogs(entries []LogEntry) {
	panic("implement me")
}

func (ms *memoryStorage) LatestLogEntry() (logEntry LogEntry, ok bool) {
	if len(ms.logEntries) == 0 {
		return go_raft.LogEntry{}, false
	}

	return ms.logEntries[len(ms.logEntries)-1], true
}
