package go_raft

type messageKind string

const (
	// Replicate log entries from leader to followers + leader heartbeats
	msgAppendEntries         = "msgAppendEntries"
	msgAppendEntriesResponse = "msgAppendEntriesResponse"

	// Leader election voting
	msgVoteFor         = "msgVoteFor"
	msgVoteForResponse = "msgVoteForResponse"

	// Progression of time is measured in ticks
	msgTick = "msgTick"

	// Proposal for a leader to append a new entry to the log - may or may not succeed
	msgProposal = "msgProposal"
)

type Message struct {
	// Kind of message
	Kind messageKind

	// Sender/receiver actor
	From ServerID
	To   ServerID

	// Term the message belongs to (only applies to some messages)
	Term Term

	// VoteFor messages only
	LastLogTerm  Term
	LastLogIndex LogIndex

	// VoteForResponse only
	VoteGranted bool

	// AppendEntries messages only
	PreviousLogTerm  Term
	PreviousLogIndex LogIndex
	LeaderCommit     LogIndex
	LogEntries       []LogEntry

	// AppendEntriesResponse messages only
	Success    bool
	MatchIndex LogIndex

	// Proposal messages only
	ProposalPayload string
}

func NewMessageVoteFor(to ServerID, from ServerID, term Term, lastLogIndex LogIndex, lastLogTerm Term) Message {
	return Message{
		Kind:         msgVoteFor,
		From:         from,
		To:           to,
		Term:         term,
		LastLogIndex: lastLogIndex,
		LastLogTerm:  lastLogTerm,
	}
}

func NewMessageVoteForResponse(to ServerID, from ServerID, term Term, voteGranted bool) Message {
	return Message{
		Kind:        msgVoteForResponse,
		From:        from,
		To:          to,
		Term:        term,
		VoteGranted: voteGranted,
	}
}

func NewMessageAppendEntries(to ServerID, from ServerID, term Term, leaderCommit LogIndex, previousLogIndex LogIndex, previousLogTerm Term, logEntries []LogEntry) Message {
	return Message{
		Kind:             msgAppendEntries,
		From:             from,
		To:               to,
		Term:             term,
		LeaderCommit:     leaderCommit,
		PreviousLogIndex: previousLogIndex,
		PreviousLogTerm:  previousLogTerm,
		LogEntries:       logEntries,
	}
}

func NewMessageAppendEntriesResponse(to ServerID, from ServerID, term Term, success bool, matchIndex LogIndex) Message {
	return Message{
		Kind:       msgAppendEntriesResponse,
		From:       from,
		To:         to,
		Term:       term,
		Success:    success,
		MatchIndex: matchIndex,
	}
}

func NewMessageTick(to ServerID, from ServerID) Message {
	return Message{
		Kind: msgTick,
		From: from,
		To:   to,
	}
}

func NewMessageProposal(to ServerID, from ServerID, payload string) Message {
	return Message{
		Kind:            msgProposal,
		From:            from,
		To:              to,
		ProposalPayload: payload,
	}
}

// HasTerm returns true if the message is associated with a term, and thus is only valid within
// that term.
func (m *Message) HasTerm() bool {
	switch m.Kind {
	case msgTick:
		return false
	case msgProposal:
		return false
	default:
		return true
	}
}
