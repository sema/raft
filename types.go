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
