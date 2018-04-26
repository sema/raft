package go_raft

type commandKind string

const (
	cmdAppendEntries = "cmdAppendEntries"
	cmdAppendEntriesResponse = "cmdAppendEntriesResponse"

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

	Success bool
	MatchIndex LogIndex
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

func newCommandAppendEntries(to ServerID, from ServerID, term Term, leaderCommit LogIndex, previousLogIndex LogIndex, previousLogTerm Term) Command {
	return Command{
		Kind: cmdAppendEntries,
		Term:   term,
		From:     from,
		LeaderCommit: leaderCommit,
		PreviousLogIndex: previousLogIndex,
		PreviousLogTerm:  previousLogTerm,
		// TODO actual log entries
	}
}

func newCommandAppendResponseEntries(to ServerID, from ServerID, term Term, success bool, matchIndex LogIndex) Command {
	return Command{
		Kind: cmdAppendEntriesResponse,
		Term:   term,
		From:     from,
		Success: success,
		MatchIndex: matchIndex,
	}
}