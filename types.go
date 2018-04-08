package raft

type ServerState int

const (
	Init      ServerState = iota
	Leader    ServerState = iota
	Candidate ServerState = iota
	Follower  ServerState = iota
)

type Term int
type NodeName string
type LogIndex int

type LogEntry struct {
	term  Term
	index LogIndex
}

// AppendEntriesRequest contain the request payload for the AppendEntries RPC
type AppendEntriesRequest struct {
	leaderTerm   Term
	leaderID     NodeName
	leaderCommit LogIndex
	prevLogIndex LogIndex
	prevLogTerm  Term
	entries      []LogEntry
}

// AppendEntriesResponse contain the response payload for the AppendEntries RPC
type AppendEntriesResponse struct {
	success bool
	term    Term
}

// RequestVoteRequest contain the request payload for the RequestVote RPC
type RequestVoteRequest struct {
	candidateTerm Term
	candidateID   NodeName
	lastLogIndex  LogIndex
	lastLogTerm   Term
}

// RequestVoteResponse contain the response payload for the RequestVote RPC
type RequestVoteResponse struct {
	term        Term
	voteGranted bool
}
