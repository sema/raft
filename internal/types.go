package internal

import "github.com/sema/raft"

type state interface {
	Enter() (nextState raft.ServerState)
}
