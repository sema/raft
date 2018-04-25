package go_raft

type CommandResult struct {
	// TODO this should not move out of the interpreter scope
	// Triggers mode and term change if NewMode != existing.
	NewTerm Term
	NewMode interpreterMode
}

func newCommandResult() *CommandResult {
	return &CommandResult{
		NewMode: existing,
		NewTerm: 0,
	}
}

func (c *CommandResult) ChangeMode(newMode interpreterMode, newTerm Term) {
	c.NewMode = newMode
	c.NewTerm = newTerm
}
