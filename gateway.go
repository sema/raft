package go_raft

type ServerGateway interface {
	SendAppendEntriesRPC(name ServerID, request AppendEntriesRequest) AppendEntriesResponse
	SendRequestVoteRPC(name ServerID, request RequestVoteRequest) RequestVoteResponse
}

func SendAppendEntriesRPC() {
	// TODO actually check for this
	// assert request.prevLogIndex + len(request.entries) >= request.leaderCommit

}
