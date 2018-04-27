package go_raft

type MessageResult struct {
	// TODO this should not move out of the Actor scope
	// Triggers mode and term change if NewMode != ExistingMode.
	NewTerm Term
	NewMode ActorMode
}

func newMessageResult() *MessageResult {
	return &MessageResult{
		NewMode: ExistingMode,
		NewTerm: 0,
	}
}

func (c *MessageResult) ChangeMode(newMode ActorMode, newTerm Term) {
	c.NewMode = newMode
	c.NewTerm = newTerm
}
