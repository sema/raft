package actor

// VolatileStorage contains in-memory state kept only during the
// lifetime of the server. This state is reset every time the server is restarted.
type VolatileStorage struct {
	CommitIndex      LogIndex
	LastAppliedIndex LogIndex
	ServerID         ServerID
}
