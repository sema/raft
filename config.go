package raft

import "math"

type Config struct {
	// List of servers participating in the consensus
	Servers []ServerID

	// Followers/Candidates initiate a new leader election if they have not observed
	// a leader within the past (timeout+random(0..splay)) ticks.
	LeaderElectionTimeout      Tick
	LeaderElectionTimeoutSplay Tick

	// Number of ticks between each Leader heartbeat. Should be strictly less than
	// the LeaderElectionTimeout to avoid spurious leader elections.
	LeaderHeartbeatFrequency Tick
}

func (c *Config) Quorum() int {
	return int(math.Floor(float64(len(c.Servers))/float64(2))) + 1
}
