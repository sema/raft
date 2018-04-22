package go_raft

import "time"

type Config struct {
	LeaderElectionTimeout      time.Duration
	LeaderElectionTimeoutSplay time.Duration
}
