package internal

import (
	"github/sema/go-raft"
)

type stateContext struct {
	state ServerState

	persistentStorage go_raft.PersistentStorage
	volatileStorage   *VolatileStorage

	gateway   go_raft.ServerGateway
	discovery go_raft.Discovery
}

func NewStateContext(serverID go_raft.ServerID, persistentStorage go_raft.PersistentStorage, gateway go_raft.ServerGateway, discovery go_raft.Discovery) StateContext {
	volatileStorage := &VolatileStorage{
		ServerID:         serverID,
		CommitIndex:      go_raft.LogIndex(0),
		LastAppliedIndex: go_raft.LogIndex(0),
	}

	return &stateContext{
		// TODO reuse the same state objects to reduce GC churn
		state:             NewFollowerState(persistentStorage, volatileStorage, gateway),
		persistentStorage: persistentStorage,
		volatileStorage:   volatileStorage,
		gateway:           gateway,
		discovery:         discovery,
	}
}

func (c *stateContext) RequestVote(request go_raft.RequestVoteRequest) go_raft.RequestVoteResponse {
	for {
		response, newState := c.state.HandleRequestVote(request)

		if newState != nil {
			c.transitionState(newState)
		} else {
			return response
		}
	}
}

func (c *stateContext) AppendEntries(request go_raft.AppendEntriesRequest) go_raft.AppendEntriesResponse {
	for {
		response, newState := c.state.HandleAppendEntries(request)

		if newState != nil {
			c.transitionState(newState)
		} else {
			return response
		}
	}
}

func (c *stateContext) TriggerLeaderElection() {
	// Don't repeat request in this case
	newState := c.state.TriggerLeaderElection()

	if newState != nil {
		c.transitionState(newState)
	}

	// TODO need to cleanup the state change interaction as it is messy, repeats logic, and has a very non-obvious
	// twist in the logic in the TriggerLeaderElection method
}

func (c *stateContext) transitionState(newState ServerState) {
	c.state = newState
}
