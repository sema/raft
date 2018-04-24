package go_raft

type serverGatewayStub struct {
}

func NewServerGatewayStub() ServerGateway {
	return &serverGatewayStub{}
}

func (g *serverGatewayStub) SendRequestVoteResponseRPC(name ServerID, from ServerID, term Term, voteGranted bool) {
	panic("implement me")
}

func (g *serverGatewayStub) SendAppendEntriesRPC(name ServerID, request AppendEntriesRequest) AppendEntriesResponse {
	panic("implement me")
}

func (g *serverGatewayStub) SendRequestVoteRPC(to ServerID, from ServerID, term Term, lastLogIndex LogIndex, lastLogTerm Term) {
	panic("implement me")
}
