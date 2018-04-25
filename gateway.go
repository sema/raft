package go_raft

type ServerGateway interface {
	Send(to ServerID, command Command)
}

type localGateway struct {
	servers map[ServerID]Server
}

func NewLocalServerGateway(servers map[ServerID]Server) ServerGateway {
	return &localGateway{
		servers: servers,
	}
}

func (g *localGateway) Send(to ServerID, command Command) {
	server := g.servers[to]
	server.SendCommand(command)
}
