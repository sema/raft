package internal

import "github.com/sema/raft"

type Volatile struct {
	// Volatile
	commitIndex      raft.LogIndex
	lastAppliedIndex raft.LogIndex
	state            raft.ServerState
	nodeName         raft.NodeName

	// Volatile - leader
	nextIndex  raft.LogIndex
	matchIndex raft.LogIndex
}
