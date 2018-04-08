package raft

type ServerGateway interface {
	SendAppendEntriesRPC(name NodeName, request AppendEntriesRequest) AppendEntriesResponse
	SendRequestVoteRPC(name NodeName, request RequestVoteRequest) RequestVoteResponse
}

func SendAppendEntriesRPC() {
	// TODO actually check for this
	// assert request.prevLogIndex + len(request.entries) >= request.leaderCommit

}
