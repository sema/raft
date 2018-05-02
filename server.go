package raft

import (
	"errors"
	"log"
	"time"
)

const tickDuration = 10 * time.Millisecond
const inboxBufferSize = 100

type Term uint64
type ServerID string
type Tick uint64

type Server interface {
	Start()
	Stop()
	SendMessage(Message) error
	CurrentStateName() string

	Log(index LogIndex) (entry LogEntry, ok bool)
	CommitIndex() LogIndex
	Age() Tick
}

type server struct {
	actor    Actor
	ticker   *time.Ticker
	inbox    chan Message
	serverID ServerID
	stop     chan bool
	done     chan bool
}

func NewServer(serverID ServerID, storage Storage, gateway ServerGateway, config Config) Server {
	actor := NewActor(serverID, storage, gateway, config)

	return &server{
		actor:    actor,
		ticker:   time.NewTicker(tickDuration),
		inbox:    make(chan Message, inboxBufferSize),
		serverID: serverID,
		stop:     make(chan bool, 1),
		done:     make(chan bool, 1),
	}
}

// Start blocks and processes messages (from SendMessage) and ticks until the Stop method is called.
func (s *server) Start() {
	for {
		select {
		case <-s.ticker.C:
			s.actor.Process(NewMessageTick(s.serverID, s.serverID))
		case message := <-s.inbox:
			s.actor.Process(message)
		case <-s.stop:
			s.done <- true
			return
		}
	}
}

// Stop stops the server, causing any invocation of Start to return. The Stop method blocks until the server has stopped
// processing any additional messages/ticks.
func (s *server) Stop() {
	s.stop <- true
	<-s.done
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

func (s *server) Log(index LogIndex) (entry LogEntry, ok bool) {
	return s.actor.Log(index)
}

func (s *server) CommitIndex() LogIndex {
	return s.actor.CommitIndex()
}

func (s *server) Age() Tick {
	return s.actor.Age()
}
