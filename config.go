package go_raft

import "time"

type Config struct {
	TickFrequency              time.Duration
	LeaderElectionTimeout      Tick
	LeaderElectionTimeoutSplay Tick
}

