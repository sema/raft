package go_raft

import "log"

type Server interface {
	Run()

	SendCommand(Command)

	CurrentStateName() string
}

type server struct {
	interpreter interpreter
	heartbeatMonitor HeartbeatMonitor

	persistentStorage PersistentStorage  // TODO remove this when heartbeat is gone
	commandQueue chan Command
}

func NewServer(serverID ServerID, storage PersistentStorage, gateway ServerGateway, discovery ServerDiscovery, config Config) Server {
	interpreter := newInterpreter(serverID, storage, gateway, discovery)

	heartbeatMonitor := NewHeartbeatMonitor(config.LeaderElectionTimeout, config.LeaderElectionTimeoutSplay)

	// TODO send reset signal to monitor on AppendEntries calls
	// c.leaderElectionTimeoutTimer.Reset(c.LeaderElectionTimeout)

	return &server{
		interpreter: interpreter,
		heartbeatMonitor: heartbeatMonitor,
		commandQueue: make(chan Command),
		persistentStorage: storage,
	}
}

func (s *server) Run() {
	// TODO replace with a ticker model as in the etcd/raft impl?
	go s.heartbeatMonitor.Run()

	go func() {
		for {
			select {
			case <-s.heartbeatMonitor.Signal():
				s.SendCommand(Command{
					Kind: cmdStartLeaderElection,
					Term: s.persistentStorage.CurrentTerm(),
				})
			}
		}
	}()

	go func() {
		for {
			command := <-s.commandQueue
			s.interpreter.Execute(command)
		}
	}()
}

func (s *server) SendCommand(command Command) {
	log.Printf("Adding command %s to queue", command.Kind)
	s.commandQueue <- command
}

func (s *server) CurrentStateName() string {
	return s.interpreter.ModeName()
}
