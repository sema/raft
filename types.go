package go_raft

type Term int
type ServerID string
type LogIndex int

func MaxLogIndex(v1 LogIndex, v2 LogIndex) LogIndex {
	if v1 > v2 {
		return v1
	}

	return v2
}

func MinLogIndex(v1 LogIndex, v2 LogIndex) LogIndex {
	if v1 < v2 {
		return v1
	}

	return v2
}

type LogEntry struct {
	Term  Term
	Index LogIndex
	Payload string
}

func NewLogEntry(term Term, index LogIndex, payload string) LogEntry {
	return LogEntry{
		Term:    term,
		Index:   index,
		Payload: payload,
	}
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
	LatestLogEntry() (logEntry LogEntry)
	AppendLog(payload string)
	LogLength() int
	MergeLogs([]LogEntry)
}

type actorModeStrategy interface {
	Name() string

	PreExecuteModeChange(message Message) (newMode ActorMode, newTerm Term)
	Process(message Message) (result *MessageResult)

	Enter()
	Exit()
}

type ActorMode int

const (
	FollowerMode ActorMode = iota
	CandidateMode
	LeaderMode

	ExistingMode // special mode to signal a no-op change to modes
)

/*
type Message interface {
	PreExecuteModeChange(mode ActorMode) (ActorMode, Term, bool)
	Process(mode ActorMode) *MessageResult

	Term() Term
}
*/
