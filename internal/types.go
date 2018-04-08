package internal

import "github.com/sema/raft"

type StateImpl interface {
	Enter() (nextState raft.ServerState)
}

type RPCRequest struct {
	Request  interface{}
	Response chan interface{}
}
