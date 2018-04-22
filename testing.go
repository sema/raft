package go_raft

type serverGatewayStub struct {
}

func NewServerGatewayStub() ServerGateway {
	return &serverGatewayStub{}
}

func (g *serverGatewayStub) SendAppendEntriesRPC(name ServerID, request AppendEntriesRequest) AppendEntriesResponse {
	panic("implement me")
}

func (g *serverGatewayStub) SendRequestVoteRPC(name ServerID, request RequestVoteRequest) RequestVoteResponse {
	panic("implement me")
}
