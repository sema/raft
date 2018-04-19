package go_raft

// PersistentStorage defines the interface for any persistent persistentStorage required by the Raft protocol.
type PersistentStorage interface {
	CurrentTerm() Term
	SetCurrentTerm(newTerm Term)

	VotedFor() NodeName
	ClearVotedFor()
	SetVotedForIfUnset(votedFor NodeName)

	Log(index LogIndex) (logEntry LogEntry, ok bool)
	LatestLogEntry() (logEntry LogEntry, ok bool)
	AppendLog(entry LogEntry)

	// Merges entries into the current log, overwriting any entries with overlapping indexes but different terms
	MergeLogs(entries []LogEntry)
}

// memoryStorage implements the PersistentStorage interface using a memory back persistentStorage. Should only be used for testing!
type memoryStorage struct {
	currentTerm Term
	votedFor    NodeName
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

func (ms *memoryStorage) VotedFor() NodeName {
	return ms.votedFor
}

func (ms *memoryStorage) SetVotedForIfUnset(votedFor NodeName) {
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

func (ms *memoryStorage) LatestLogEntry() (logEntry LogEntry, ok bool) {
	if len(ms.logEntries) == 0 {
		return LogEntry{}, false
	}

	return ms.logEntries[len(ms.logEntries)-1], true
}
