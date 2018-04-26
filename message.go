package go_raft

type messageKind string

const (
	msgAppendEntries         = "msgAppendEntries"
	msgAppendEntriesResponse = "msgAppendEntriesResponse"
	msgVoteFor               = "msgVoteFor"
	msgVoteForResponse       = "msgVoteForResponse"
	msgTick                  = "msgTick"
)

// TODO cleanup fields
type Message struct {
	Kind messageKind

	Term Term

	LastLogTerm  Term
	LastLogIndex LogIndex

	PreviousLogTerm  Term
	PreviousLogIndex LogIndex

	LeaderCommit LogIndex

	VoteGranted bool

	From ServerID

	Success    bool
	MatchIndex LogIndex
}

func NewMessageVoteFor(to ServerID, from ServerID, term Term, lastLogIndex LogIndex, lastLogTerm Term) Message {
	return Message{
		Kind:         msgVoteFor,
		From:         from,
		Term:         term,
		LastLogIndex: lastLogIndex,
		LastLogTerm:  lastLogTerm,
	}
}

func NewMessageVoteForResponse(to ServerID, from ServerID, term Term, voteGranted bool) Message {
	return Message{
		Kind:        msgVoteForResponse,
		Term:        term,
		VoteGranted: voteGranted,
		From:        from,
	}
}

func newMessageAppendEntries(to ServerID, from ServerID, term Term, leaderCommit LogIndex, previousLogIndex LogIndex, previousLogTerm Term) Message {
	return Message{
		Kind:             msgAppendEntries,
		Term:             term,
		From:             from,
		LeaderCommit:     leaderCommit,
		PreviousLogIndex: previousLogIndex,
		PreviousLogTerm:  previousLogTerm,
		// TODO actual log entries
	}
}

func newMessageAppendResponseEntries(to ServerID, from ServerID, term Term, success bool, matchIndex LogIndex) Message {
	return Message{
		Kind:       msgAppendEntriesResponse,
		Term:       term,
		From:       from,
		Success:    success,
		MatchIndex: matchIndex,
	}
}
