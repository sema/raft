package go_raft

type ServerGateway interface {
	SendAppendEntriesRPC(to ServerID, request AppendEntriesRequest) AppendEntriesResponse
	SendRequestVoteRPC(to ServerID, from ServerID, term Term, lastLogIndex LogIndex, lastLogTerm Term)
	SendRequestVoteResponseRPC(to ServerID, from ServerID, term Term, voteGranted bool)
}

type localGateway struct {
	servers map[ServerID]Server
}

func NewLocalServerGateway(servers map[ServerID]Server) ServerGateway {
	return &localGateway{
		servers: servers,
	}
}

func (g *localGateway) SendAppendEntriesRPC(to ServerID, request AppendEntriesRequest) AppendEntriesResponse {
	// TODO actually check for this
	// assert request.prevLogIndex + len(request.entries) >= request.leaderCommit
	panic("implement me")
}

func (g *localGateway) SendRequestVoteRPC(to ServerID, from ServerID, term Term, lastLogIndex LogIndex, lastLogTerm Term) {
	server := g.servers[to]
	server.SendCommand(Command{
		Kind: cmdVoteFor,
		From: from,
		Term: term,
		LastLogIndex: lastLogIndex,
		LastLogTerm: lastLogTerm,
	})
}

func (g *localGateway) SendRequestVoteResponseRPC(to ServerID, from ServerID, term Term, voteGranted bool) {
	server := g.servers[to]
	server.SendCommand(Command{
		Kind: cmdVoteForResponse,
		Term: term,
		VoteGranted: voteGranted,
		From: from,
	})
}
