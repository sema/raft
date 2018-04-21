package go_raft

// TODO determine if we need to do locking at this level, or if we can push it down
type threadsafeStateContext struct {
	wrapped stateContext
	mutex   chan bool
}

func NewThreadsafeStateContext(context stateContext) stateContext {
	return &threadsafeStateContext{
		wrapped: context,
		mutex:   make(chan bool, 1),
	}
}

func (c *threadsafeStateContext) RequestVote(request RequestVoteRequest) RequestVoteResponse {
	c.takeLock()
	defer c.releaseLock()

	return c.wrapped.RequestVote(request)
}

func (c *threadsafeStateContext) AppendEntries(request AppendEntriesRequest) AppendEntriesResponse {
	c.takeLock()
	defer c.releaseLock()

	return c.wrapped.AppendEntries(request)
}

func (c *threadsafeStateContext) TriggerLeaderElection() {
	c.takeLock()
	defer c.releaseLock()

	c.wrapped.TriggerLeaderElection()
}

func (c *threadsafeStateContext) takeLock() {
	<-c.mutex
}

func (c *threadsafeStateContext) releaseLock() {
	c.mutex <- true
}
