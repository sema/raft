package go_raft

import "log"

type Server interface {
	Run()

	RequestVote(RequestVoteRequest) RequestVoteResponse
	AppendEntries(AppendEntriesRequest) AppendEntriesResponse

	CurrentStateName() string
}

type server struct {
	stateContext     stateContext
	heartbeatMonitor HeartbeatMonitor
}

func NewServer(serverID ServerID, storage PersistentStorage, gateway ServerGateway, discovery Discovery, config Config) Server {
	context := newStateContext(serverID, storage, gateway, discovery)
	context = newThreadsafeStateContext(context)

	heartbeatMonitor := NewHeartbeatMonitor(config.LeaderElectionTimeout, config.LeaderElectionTimeoutSplay)

	// TODO send reset signal to monitor on AppendEntries calls
	// c.leaderElectionTimeoutTimer.Reset(c.LeaderElectionTimeout)

	return &server{
		stateContext:     context,
		heartbeatMonitor: heartbeatMonitor,
	}
}

func (s *server) Run() {
	go s.heartbeatMonitor.Run()

	for {
		select {
		case <-s.heartbeatMonitor.Signal():
			s.triggerLeaderElection()
		}
	}
}

func (s *server) RequestVote(request RequestVoteRequest) RequestVoteResponse {
	log.Print("Event RequestVote")
	return s.stateContext.RequestVote(request)
}

func (s *server) AppendEntries(request AppendEntriesRequest) AppendEntriesResponse {
	log.Print("Event AppendEntries")
	return s.stateContext.AppendEntries(request)
}

func (s *server) triggerLeaderElection() {
	log.Print("Event TriggerLeaderElection")
	s.stateContext.TriggerLeaderElection()
}

func (s *server) CurrentStateName() string {
	return s.stateContext.CurrentStateName()
}
