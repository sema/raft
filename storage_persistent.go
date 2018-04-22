package go_raft

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
		// Don't clear vote if we are setting the same term multiple times, which might occur in certain edge cases
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
	// TODO this need to be atomic?
	if index == 0 {
		return LogEntry{}, false
	}

	if len(ms.logEntries) < int(index) {
		return LogEntry{}, false
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

func (ms *memoryStorage) LatestLogEntry() (logEntry LogEntry) {
	if len(ms.logEntries) == 0 {
		return LogEntry{
			Term:  0,
			Index: 0,
		}
	}

	return ms.logEntries[len(ms.logEntries)-1]
}
