package raft

import (
	"log"
	"time"
)

type Server interface {
	Run()
	SendMessage(Message)
	CurrentStateName() string
}

type server struct {
	interpreter Actor
	ticker      *time.Ticker
	inbox       chan Message
}

func NewServer(serverID ServerID, storage Storage, gateway ServerGateway, config Config) Server {
	interpreter := NewActor(serverID, storage, gateway, config)

	return &server{
		interpreter: interpreter,
		ticker:      time.NewTicker(10 * time.Millisecond),
		inbox:       make(chan Message, 100), // TODO handle deadlocks caused by channel overflow
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
			s.interpreter.Process(message)
		}
	}()
}

func (s *server) SendMessage(message Message) {
	log.Printf("Adding message %s to queue", message.Kind)
	s.inbox <- message
}

func (s *server) CurrentStateName() string {
	return s.interpreter.ModeName()
}

func (s *server) tick() {
	s.SendMessage(Message{
		Kind: msgTick,
	})
}
