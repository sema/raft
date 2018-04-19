package internal

import "github/sema/go-raft"

// VolatileStorage contains in-memory state kept only during the
// lifetime of the server. This state is reset every time the server is restarted.
type VolatileStorage struct {
	// Volatile
	commitIndex      go_raft.LogIndex
	lastAppliedIndex go_raft.LogIndex
	state            go_raft.ServerState
	nodeName         go_raft.NodeName

	// Volatile - leader
	nextIndex  go_raft.LogIndex
	matchIndex go_raft.LogIndex
}
