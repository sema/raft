package raft

import (
	"log"
	"time"
)

type Server interface {
	Run()
	SendMessage(Message)
	CurrentStateName() string

	Log(index LogIndex) (entry LogEntry, ok bool)
	CommitIndex() LogIndex
	Age() Tick
}

type server struct {
	actor   Actor
	ticker  *time.Ticker
	inbox   chan Message
}

func NewServer(serverID ServerID, storage Storage, gateway ServerGateway, config Config) Server {
	actor := NewActor(serverID, storage, gateway, config)

	return &server{
		actor:   actor,
		ticker:  time.NewTicker(10 * time.Millisecond),
		inbox:   make(chan Message, 100), // TODO handle deadlocks caused by channel overflow
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
			message := <-s.inbox
			s.actor.Process(message)
		}
	}()
}

func (s *server) SendMessage(message Message) {
	log.Printf("Adding message %s to queue", message.Kind)
	s.inbox <- message
}

func (s *server) CurrentStateName() string {
	return s.actor.ModeName()
}

func (s *server) tick() {
	s.SendMessage(Message{
		Kind: msgTick,
	})
}

func (s *server) Log(index LogIndex) (entry LogEntry, ok bool) {
	return s.actor.Log(index)
}

func (s *server) CommitIndex() LogIndex {
	return s.actor.CommitIndex()
}

func (s *server) Age() Tick {
	return s.actor.Age()
}