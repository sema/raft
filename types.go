package go_raft

//type ServerState int

/*const (
	Init      ServerState = iota
	Leader    ServerState = iota
	Candidate ServerState = iota
	Follower  ServerState = iota
)*/

type Term int
type NodeName string
type LogIndex int

type LogEntry struct {
	Term  Term
	Index LogIndex
}

// AppendEntriesRequest contain the request payload for the AppendEntries RPC
type AppendEntriesRequest struct {
	LeaderTerm   Term
	LeaderID     NodeName
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
	CandidateID   NodeName
	LastLogIndex  LogIndex
	LastLogTerm   Term
}

// RequestVoteResponse contain the response payload for the RequestVote RPC
type RequestVoteResponse struct {
	Term        Term
	VoteGranted bool
}
