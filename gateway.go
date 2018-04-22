package go_raft

type ServerGateway interface {
	SendAppendEntriesRPC(name ServerID, request AppendEntriesRequest) AppendEntriesResponse
	SendRequestVoteRPC(name ServerID, request RequestVoteRequest) RequestVoteResponse
}

type localGateway struct {
	servers map[ServerID]Server
}

func NewLocalServerGateway(servers map[ServerID]Server) ServerGateway {
	return &localGateway{
		servers: servers,
	}
}

func (g *localGateway) SendAppendEntriesRPC(name ServerID, request AppendEntriesRequest) AppendEntriesResponse {
	// TODO actually check for this
	// assert request.prevLogIndex + len(request.entries) >= request.leaderCommit
	panic("implement me")
}

func (g *localGateway) SendRequestVoteRPC(name ServerID, request RequestVoteRequest) RequestVoteResponse {
	server := g.servers[name]
	return server.RequestVote(request)
}
