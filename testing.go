package go_raft

type serverGatewayStub struct {
}

func (g *serverGatewayStub) Send(to ServerID, command Command) {
	panic("implement me")
}

func NewServerGatewayStub() ServerGateway {
	return &serverGatewayStub{}
}

