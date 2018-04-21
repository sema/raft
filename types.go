package go_raft

type Term int
type ServerID string
type LogIndex int

type LogEntry struct {
	Term  Term
	Index LogIndex
}

// AppendEntriesRequest contain the request payload for the AppendEntries RPC
type AppendEntriesRequest struct {
	LeaderTerm   Term
	LeaderID     ServerID
	LeaderCommit LogIndex
	PrevLogIndex LogIndex
	PrevLogTerm  Term
	Entries      []LogEntry
}

// AppendEntriesResponse contain the response payload for the AppendEntries RPC
type AppendEntriesResponse struct {
	Success bool
	Term    Term
}

// RequestVoteRequest contain the request payload for the RequestVote RPC
type RequestVoteRequest struct {
	CandidateTerm Term
	CandidateID   ServerID
	LastLogIndex  LogIndex
	LastLogTerm   Term
}

// RequestVoteResponse contain the response payload for the RequestVote RPC
type RequestVoteResponse struct {
	Term        Term
	VoteGranted bool
}

// PersistentStorage defines the interface for any persistent persistentStorage required by the Raft protocol.
type PersistentStorage interface {
	CurrentTerm() Term
	SetCurrentTerm(newTerm Term)

	VotedFor() ServerID
	ClearVotedFor()
	SetVotedForIfUnset(votedFor ServerID)

	Log(index LogIndex) (logEntry LogEntry, ok bool)
	LatestLogEntry() (logEntry LogEntry, ok bool)
	AppendLog(entry LogEntry)

	// Merges entries into the current log, overwriting any entries with overlapping indexes but different terms
	MergeLogs(entries []LogEntry)
}

type serverState interface {
	Name() string

	HandleRequestVote(RequestVoteRequest) (response RequestVoteResponse, newState serverState)
	HandleAppendEntries(AppendEntriesRequest) (response AppendEntriesResponse, newState serverState)
	TriggerLeaderElection() (newState serverState)
}
