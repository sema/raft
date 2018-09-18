package grpcserver

import "github.com/sema/raft/pkg/actor"

// DiscoveryConfig represents the listening and reachable address of a server and is typically statically configured.
type DiscoveryConfig struct {
	serverID      actor.ServerID
	addressLocal  string
	addressRemote string
}
