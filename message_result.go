package go_raft

type MessageResult struct {
	// TODO this should not move out of the actor scope
	// Triggers mode and term change if NewMode != existing.
	NewTerm Term
	NewMode actorMode
}

func newMessageResult() *MessageResult {
	return &MessageResult{
		NewMode: existing,
		NewTerm: 0,
	}
}

func (c *MessageResult) ChangeMode(newMode actorMode, newTerm Term) {
	c.NewMode = newMode
	c.NewTerm = newTerm
}
