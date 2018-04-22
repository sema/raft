package go_raft

type Server interface {
	Run()
	RequestVote(RequestVoteRequest) RequestVoteResponse
	AppendEntries(AppendEntriesRequest) AppendEntriesResponse
}

type server struct {
	stateContext     stateContext
	heartbeatMonitor HeartbeatMonitor
}

func NewServer(serverID ServerID, storage PersistentStorage, gateway ServerGateway, discovery Discovery, config Config) Server {
	context := newStateContext(serverID, storage, gateway, discovery)
	context = NewThreadsafeStateContext(context)

	heartbeatMonitor := NewHeartbeatMonitor(config.leaderElectionTimeout, config.leaderElectionTimeoutSplay)

	// TODO send reset signal to monitor on AppendEntries calls
	// c.leaderElectionTimeoutTimer.Reset(c.leaderElectionTimeout)

	return &server{
		stateContext:     context,
		heartbeatMonitor: heartbeatMonitor,
	}
}

func (s *server) Run() {
	for {
		select {
		case <-s.heartbeatMonitor.Signal():
			s.stateContext.TriggerLeaderElection()
		}
	}
}

func (s *server) RequestVote(request RequestVoteRequest) RequestVoteResponse {
	return s.stateContext.RequestVote(request)
}

func (s *server) AppendEntries(request AppendEntriesRequest) AppendEntriesResponse {
	return s.stateContext.AppendEntries(request)
}
