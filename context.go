package go_raft

type stateContext interface {
	RequestVote(request RequestVoteRequest) RequestVoteResponse
	AppendEntries(request AppendEntriesRequest) AppendEntriesResponse
	TriggerLeaderElection()
}

type stateContextImpl struct {
	state serverState

	persistentStorage PersistentStorage
	volatileStorage   *VolatileStorage

	gateway   ServerGateway
	discovery Discovery
}

func newStateContext(serverID ServerID, persistentStorage PersistentStorage, gateway ServerGateway, discovery Discovery) stateContext {
	volatileStorage := &VolatileStorage{
		ServerID:         serverID,
		CommitIndex:      LogIndex(0),
		LastAppliedIndex: LogIndex(0),
	}

	return &stateContextImpl{
		state:             newFollowerState(persistentStorage, volatileStorage, gateway, discovery),
		persistentStorage: persistentStorage,
		volatileStorage:   volatileStorage,
		gateway:           gateway,
		discovery:         discovery,
	}
}

func (c *stateContextImpl) RequestVote(request RequestVoteRequest) RequestVoteResponse {
	for {
		response, newState := c.state.HandleRequestVote(request)

		if newState != nil {
			c.transitionState(newState)
		} else {
			return response
		}
	}
}

func (c *stateContextImpl) AppendEntries(request AppendEntriesRequest) AppendEntriesResponse {
	for {
		response, newState := c.state.HandleAppendEntries(request)

		if newState != nil {
			c.transitionState(newState)
		} else {
			return response
		}
	}
}

func (c *stateContextImpl) TriggerLeaderElection() {
	// Don't repeat request in this case
	newState := c.state.TriggerLeaderElection()

	if newState != nil {
		c.transitionState(newState)
	}

	// TODO need to cleanup the state change interaction as it is messy, repeats logic, and has a very non-obvious
	// twist in the logic in the TriggerLeaderElection method
}

func (c *stateContextImpl) transitionState(newState serverState) {
	c.state = newState
}
