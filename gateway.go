package go_raft

type ServerGateway interface {
	SendAppendEntriesRPC(name ServerID, request AppendEntriesRequest) AppendEntriesResponse
	SendRequestVoteRPC(name ServerID, request RequestVoteRequest) RequestVoteResponse
}

type gatewayImpl struct {
}

func NewServerGateway() ServerGateway {
	return &gatewayImpl{}
}

func (*gatewayImpl) SendAppendEntriesRPC(name ServerID, request AppendEntriesRequest) AppendEntriesResponse {
	// TODO actually check for this
	// assert request.prevLogIndex + len(request.entries) >= request.leaderCommit
	panic("implement me")
}

func (*gatewayImpl) SendRequestVoteRPC(name ServerID, request RequestVoteRequest) RequestVoteResponse {
	panic("implement me")
}
