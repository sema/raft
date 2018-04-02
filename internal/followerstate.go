package internal

import (
	"time"

	"github.com/sema/raft"
)

type followerState struct {
}

func newFollowerState() State {
	return &followerState{}
}

func (f *followerState) Enter() (nextState raft.ServerState) {
	leaderElectionTimeout := time.NewTimer(1 * time.Second)

	for {
		select {
		case <-leaderElectionTimeout.C:
			// No RPCs observed from leaders or other candidates, transition to Candidate state
			return raft.Candidate

		}
	}

}
