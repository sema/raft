package server

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/sema/raft/pkg/actor"
)

const tickDuration = 10 * time.Millisecond
const inboxBufferSize = 100
const outboxBufferSize = 100

type Server interface {
	Start()
	Stop()

	SendMessage(actor.Message) error
	Outbox() <-chan actor.Message

	Log(index actor.LogIndex) (entry actor.LogEntry, ok bool)
	CommitIndex() actor.LogIndex

	CurrentStateName() string
	Age() actor.Tick
}

type server struct {
	actor    actor.Actor
	ticker   *time.Ticker
	inbox    chan actor.Message
	outbox   chan actor.Message
	serverID actor.ServerID

	quitOnce sync.Once
	quit     chan bool
	doneOnce sync.Once
	done     chan bool
}

func NewServer(serverID actor.ServerID, storage actor.Storage, config actor.Config) Server {
	return &server{
		actor:    actor.NewActor(serverID, storage, config),
		ticker:   time.NewTicker(tickDuration),
		inbox:    make(chan actor.Message, inboxBufferSize),
		outbox:   make(chan actor.Message, outboxBufferSize),
		serverID: serverID,
		quit:     make(chan bool, 1),
		done:     make(chan bool, 1),
	}
}

// Start blocks and processes messages (from SendMessage) and ticks until the Stop method is called.
func (s *server) Start() {
	for {
		var messages []actor.Message

		select {
		case <-s.ticker.C:
			messages = s.actor.Process(actor.NewMessageTick(s.serverID, s.serverID))
		case message := <-s.inbox:
			messages = s.actor.Process(message)
		case <-s.quit:
			s.doneOnce.Do(func() {
				close(s.done)
			})
			return
		}

		for _, message := range messages {
			select {
			case s.outbox <- message:
				// message added to outbox
			default:
				log.Printf("Unable to add message %s to outbox as outbox is full", message.Kind)
			}
		}
	}
}

// Stop stops the server. The Stop method blocks until the server has stopped processing any additional messages/ticks.
// It is not possible to start a stopped server again.
func (s *server) Stop() {
	s.quitOnce.Do(func() {
		close(s.quit)
	})

	<-s.done
}

// SendMessage adds a message to the server's inbox for later processing. Raises an error
// if unable to add the message to the queue, for example, due to the inbox being full.
func (s *server) SendMessage(message actor.Message) error {
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

func (s *server) Log(index actor.LogIndex) (entry actor.LogEntry, ok bool) {
	return s.actor.Log(index)
}

func (s *server) CommitIndex() actor.LogIndex {
	return s.actor.CommitIndex()
}

func (s *server) Age() actor.Tick {
	return s.actor.Age()
}

func (s *server) Outbox() <-chan actor.Message {
	return s.outbox
}
