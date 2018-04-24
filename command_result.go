package go_raft

type CommandResult struct {
	Success bool
	Term    Term

	NewTerm Term
	NewMode interpreterMode
}

func newCommandResult(success bool, term Term) *CommandResult {
	return &CommandResult{
		Success: success,
		Term:    term,
		NewTerm: 0,
		NewMode: existing,
	}
}

func (c *CommandResult) ChangeMode(newMode interpreterMode, newTerm Term) {
	c.NewMode = newMode
	c.NewTerm = newTerm
}

