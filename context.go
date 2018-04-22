package go_raft

import "log"

type stateContext interface {
	RequestVote(request RequestVoteRequest) RequestVoteResponse
	AppendEntries(request AppendEntriesRequest) AppendEntriesResponse
	TriggerLeaderElection()

	TransitionStateIf(newState serverState, term Term)

	CurrentStateName() string
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

	context := &stateContextImpl{
		persistentStorage: persistentStorage,
		volatileStorage:   volatileStorage,
		gateway:           gateway,
		discovery:         discovery,
	}

	context.state = newFollowerState(persistentStorage, volatileStorage, gateway, discovery, context)

	return context
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

func (c *stateContextImpl) TransitionStateIf(newState serverState, term Term) {
	// TODO cleanup this mess
	if c.persistentStorage.CurrentTerm() == term {
		c.transitionState(newState)
	}
}

func (c *stateContextImpl) CurrentStateName() string {
	return c.state.Name()
}

func (c *stateContextImpl) transitionState(newState serverState) {
	log.Printf("Changing state %s -> %s", c.state.Name(), newState.Name())
	c.state.Exit()
	c.state = newState
	c.state.Enter()
}
