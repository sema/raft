package raft

//go:generate mockgen -destination=mocks/mock_gateway.go -source=gateway.go

type ServerGateway interface {
	Send(to ServerID, message Message)
}

type localGateway struct {
	servers map[ServerID]Server
}

func NewLocalServerGateway(servers map[ServerID]Server) ServerGateway {
	return &localGateway{
		servers: servers,
	}
}

func (g *localGateway) Send(to ServerID, message Message) {
	server := g.servers[to]
	server.SendMessage(message)
}
