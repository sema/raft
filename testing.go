package go_raft

type serverGatewayStub struct {
	mailbox map[ServerID][]Message
}

func (g *serverGatewayStub) Send(to ServerID, message Message) {
	_, ok := g.mailbox[to]
	if !ok {
		g.mailbox[to] = []Message{}
	}

	g.mailbox[to] = append(g.mailbox[to], message)
}

func NewServerGatewayStub() ServerGateway {
	return &serverGatewayStub{
		mailbox: make(map[ServerID][]Message),
	}
}
