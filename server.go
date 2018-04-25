package go_raft

import (
	"log"
	"time"
)

type Server interface {
	Run()
	SendCommand(Command)
	CurrentStateName() string
}

type server struct {
	interpreter      interpreter
	ticker *time.Ticker
	commandQueue chan Command
}

func NewServer(serverID ServerID, storage PersistentStorage, gateway ServerGateway, discovery ServerDiscovery, config Config) Server {
	interpreter := newInterpreter(serverID, storage, gateway, discovery)

	return &server{
		interpreter: interpreter,
		ticker: time.NewTicker(10 * time.Millisecond),
		commandQueue: make(chan Command, 100),  // TODO handle deadlocks caused by channel overflow
	}
}

func (s *server) Run() {
	// TODO place somewhere else
	go func() {
		for {
			<-s.ticker.C
			s.tick()
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

func (s *server) tick() {
	s.SendCommand(Command{
		Kind: cmdTick,
	})
}
