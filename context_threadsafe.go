package go_raft

import "sync"

// TODO determine if we need to do locking at this level, or if we can push it down
type threadsafeStateContext struct {
	wrapped stateContext
	mutex   sync.Locker
}

func newThreadsafeStateContext(context stateContext) stateContext {
	return &threadsafeStateContext{
		wrapped: context,
		mutex:   &sync.Mutex{},
	}
}

func (c *threadsafeStateContext) RequestVote(request RequestVoteRequest) RequestVoteResponse {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.wrapped.RequestVote(request)
}

func (c *threadsafeStateContext) AppendEntries(request AppendEntriesRequest) AppendEntriesResponse {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.wrapped.AppendEntries(request)
}

func (c *threadsafeStateContext) TriggerLeaderElection() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.wrapped.TriggerLeaderElection()
}

func (c *threadsafeStateContext) CurrentStateName() string {
	return c.wrapped.CurrentStateName()
}

func (c *threadsafeStateContext) TransitionStateIf(newState serverState, term Term) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.wrapped.TransitionStateIf(newState, term)
}