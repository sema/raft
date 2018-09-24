package grpcserver

import (
	"context"
	"io"
	"net"

	"google.golang.org/grpc"

	"strings"

	"log"

	pb "github.com/sema/raft/pkg/grpcserver/pb"
	"github.com/sema/raft/pkg/server"
)

// InboundGRPCServer exposes a Server through GRPC endpoints
type InboundGRPCServer struct {
	server     server.Server
	grpcServer *grpc.Server
}

// NewInboundGRPCServer returns new instance of InboundGRPCServer
func NewInboundGRPCServer(server server.Server) *InboundGRPCServer {
	return &InboundGRPCServer{
		server: server,
	}
}

func (s *InboundGRPCServer) GetStatus(context.Context, *pb.StatusRequest) (*pb.StatusResponse, error) {
	return &pb.StatusResponse{
		State: s.server.CurrentStateName(),
	}, nil
}

func (s *InboundGRPCServer) SendMessages(stream pb.SRServer_SendMessagesServer) error {
	for {
		message, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.SendMessagesResponse{})
		}
		if err != nil {
			return err
		}

		err = s.server.SendMessage(AsNativeMessage(message))
		if err != nil {
			return err
		}
	}
}

// Serve starts a GRPC server listening for incoming messages
//
// The serverAddr accepts unix sockets (unix:file.socket) or tcp/ip (tcp::8000, tcp:127.0.0.1:0)
func (s *InboundGRPCServer) Serve(serverAddr string) {
	parts := strings.SplitN(serverAddr, ":", 2)
	if len(parts) != 2 {
		log.Panicf("Invalid server address %s", serverAddr)
	}

	network := parts[0]
	address := parts[1]

	lis, err := net.Listen(network, address)
	if err != nil {
		log.Panicf("Unable to listen for incoming connections: %s", err)
	}

	s.grpcServer = grpc.NewServer()
	pb.RegisterSRServerServer(s.grpcServer, s)

	err = s.grpcServer.Serve(lis)
	if err != nil {
		log.Panicf("Server stopped listening for incoming connections: %s", err)
	}
}

func (s *InboundGRPCServer) Stop() {
	log.Printf("Stopping inbound GRPC server")
	s.grpcServer.Stop()
}
