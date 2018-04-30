package go_raft

type messageKind string

const (
	msgAppendEntries         = "msgAppendEntries"
	msgAppendEntriesResponse = "msgAppendEntriesResponse"
	msgVoteFor               = "msgVoteFor"
	msgVoteForResponse       = "msgVoteForResponse"
	msgTick                  = "msgTick"
	msgProposal              = "msgProposal"
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

	LogEntries []LogEntry

	ProposalPayload string
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

func NewMessageAppendEntries(to ServerID, from ServerID, term Term, leaderCommit LogIndex, previousLogIndex LogIndex, previousLogTerm Term, logEntries []LogEntry) Message {
	return Message{
		Kind:             msgAppendEntries,
		Term:             term,
		From:             from,
		LeaderCommit:     leaderCommit,
		PreviousLogIndex: previousLogIndex,
		PreviousLogTerm:  previousLogTerm,
		LogEntries:       logEntries,
	}
}

func NewMessageAppendEntriesResponse(to ServerID, from ServerID, term Term, success bool, matchIndex LogIndex) Message {
	return Message{
		Kind:       msgAppendEntriesResponse,
		Term:       term,
		From:       from,
		Success:    success,
		MatchIndex: matchIndex,
	}
}

func NewMessageTick(to ServerID, from ServerID) Message {
	return Message{
		Kind: msgTick,
		From: from,
	}
}

func NewMessageProposal(to ServerID, from ServerID, payload string) Message {
	return Message{
		Kind:            msgProposal,
		From:            from,
		ProposalPayload: payload,
	}
}
