package grpcserver

import (
	"errors"
	"log"

	"google.golang.org/grpc"

	"context"

	"fmt"

	"sync"

	"github.com/sema/raft/pkg/actor"
	pb "github.com/sema/raft/pkg/grpcserver/pb"
	"github.com/sema/raft/pkg/server"
)

// OutboundGRPCServer sends outbound messages from the outbox encoding using GRPC
type OutboundGRPCServer struct {
	source    server.Server
	discovery map[actor.ServerID]DiscoveryConfig

	streams map[actor.ServerID]pb.SRServer_SendMessagesClient

	quitOnce sync.Once
	quit     chan bool
	doneOnce sync.Once
	done     chan bool
}

// NewOutboundGRPCServer returns new instance of OutboundGRPCServer
func NewOutboundGRPCServer(server server.Server, discovery map[actor.ServerID]DiscoveryConfig) *OutboundGRPCServer {
	return &OutboundGRPCServer{
		source:    server,
		discovery: discovery,
		streams:   make(map[actor.ServerID]pb.SRServer_SendMessagesClient),
		quit:      make(chan bool),
		done:      make(chan bool),
	}
}

// Serve starts the server, sending outbound GRPC requests to remote servers when new messages arrive in the local
// server's outbox.
func (s *OutboundGRPCServer) Serve() {
	for {
		select {
		case <-s.quit:
			s.doneOnce.Do(func() {
				close(s.done)
			})
			return
		case message := <-s.source.Outbox():
			log.Printf("Sending message %s to %s", message.Kind, message.To)
			err := s.sendMessage(message)
			if err != nil {
				log.Printf("Error when sending message: %s", err)
			}
		}
	}
}

func (s *OutboundGRPCServer) sendMessage(message actor.Message) error {
	stream, err := s.getOrCreateStream(message)
	if err != nil {
		return err
	}

	err = stream.Send(AsGRPCMessage(message))
	if err != nil {
		stream.CloseSend()
		delete(s.streams, message.To)
		return err
	}

	return nil
}

func (s *OutboundGRPCServer) getOrCreateStream(message actor.Message) (pb.SRServer_SendMessagesClient, error) {
	stream, ok := s.streams[message.To]
	if ok {
		return stream, nil
	}

	stream, err := s.createStream(message.To)
	if err != nil {
		return nil, err
	}

	s.streams[message.To] = stream
	return stream, nil
}

func (s *OutboundGRPCServer) createStream(serverID actor.ServerID) (pb.SRServer_SendMessagesClient, error) {
	config, ok := s.discovery[serverID]
	if !ok {
		return nil, errors.New(fmt.Sprintf(
			"Unable to connect to server %s due to missing address", serverID))
	}

	conn, err := grpc.Dial(config.AddressRemote, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	client := pb.NewSRServerClient(conn)
	return client.SendMessages(context.Background())
}

// Stop stops the server, returning after the server has been fully stopped. The server can't be started again
// after being stopped.
func (s *OutboundGRPCServer) Stop() {
	log.Printf("Stopping outbound GRPC server")

	s.quitOnce.Do(func() {
		close(s.quit)
	})
	<-s.done
}
