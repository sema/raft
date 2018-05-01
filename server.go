package raft

import (
	"time"
	"log"
	"errors"
)

const tickDuration = 10 * time.Millisecond
const inboxBufferSize = 100

type Server interface {
	Run()
	SendMessage(Message) error
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
		ticker:  time.NewTicker(tickDuration),
		inbox:   make(chan Message, inboxBufferSize),
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

// SendMessage adds a message to the server's inbox for later processing. Raises an error
// if unable to add the message to the queue, for example, due to the inbox being full.
func (s *server) SendMessage(message Message) error {
	select {
	case s.inbox <- message:
		log.Printf("Added message %s to inbox", message.Kind)
		return nil
	default:
		log.Printf("Unable to add message %s to inbox as inbox is full", message.Kind)
		return errors.New("unable to add message to inbox as inbox is full")
	}
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