package go_raft

type serverGatewayStub struct {
}

func (g *serverGatewayStub) Send(to ServerID, message Message) {
	panic("implement me")
}

func NewServerGatewayStub() ServerGateway {
	return &serverGatewayStub{}
}
