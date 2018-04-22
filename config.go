package go_raft

import "time"

type Config struct {
	leaderElectionTimeout      time.Duration
	leaderElectionTimeoutSplay time.Duration
}
