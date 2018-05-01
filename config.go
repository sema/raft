package raft

type Config struct {
	// Followers/Candidates initiate a new leader election if they have not observed
	// a leader within the past (timeout+random(0..splay)) ticks.
	LeaderElectionTimeout      Tick
	LeaderElectionTimeoutSplay Tick

	// Number of ticks between each Leader heartbeat. Should be strictly less than
	// the LeaderElectionTimeout to avoid spurious leader elections.
	LeaderHeartbeatFrequency Tick
}
