package go_raft

type commandKind string

const (
	cmdAppendEntries = "cmdAppendEntries"
	cmdVoteFor = "cmdVoteFor"
	cmdVoteForResponse = "cmdVoteForResponse"

	cmdTick = "cmdTick"
)

type Command struct {
	Kind commandKind

	Term Term

	LastLogTerm Term
	LastLogIndex LogIndex

	PreviousLogTerm Term
	PreviousLogIndex LogIndex

	LeaderCommit LogIndex

	VoteGranted bool

	From ServerID
}

func NewCommandVoteFor(to ServerID, from ServerID, term Term, lastLogIndex LogIndex, lastLogTerm Term) Command {
	return Command{
		Kind: cmdVoteFor,
		From: from,
		Term: term,
		LastLogIndex: lastLogIndex,
		LastLogTerm: lastLogTerm,
	}
}

func NewCommandVoteForResponse(to ServerID, from ServerID, term Term, voteGranted bool) Command {
	return Command{
		Kind: cmdVoteForResponse,
		Term: term,
		VoteGranted: voteGranted,
		From: from,
	}
}
