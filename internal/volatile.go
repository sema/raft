package internal

import "github/sema/go-raft"

// VolatileStorage contains in-memory state kept only during the
// lifetime of the server. This state is reset every time the server is restarted.
type VolatileStorage struct {
	// Volatile
	CommitIndex      go_raft.LogIndex
	LastAppliedIndex go_raft.LogIndex
	ServerID go_raft.NodeName

	// Volatile - leader
	NextIndex  go_raft.LogIndex
	MatchIndex go_raft.LogIndex
}
