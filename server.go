package go_raft

import "github/sema/go-raft/internal"

type Server interface {
	Run()
	RequestVote(RequestVoteRequest) RequestVoteResponse
	AppendEntries(AppendEntriesRequest) AppendEntriesResponse
}

type server struct {
	stateContext     internal.StateContext
	heartbeatMonitor internal.HeartbeatMonitor
}

func NewServer(serverID ServerID, storage PersistentStorage, gateway ServerGateway, discovery Discovery) Server {
	context := internal.NewStateContext(serverID, storage, gateway, discovery)
	context = internal.NewThreadsafeStateContext(context)

	heartbeatMonitor := internal.NewHeartbeatMonitor()

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
