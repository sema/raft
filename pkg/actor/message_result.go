package actor

type MessageResult struct {
	// Triggers mode and term change if NewMode != ExistingMode.
	NewTerm     Term
	NewMode     ActorMode
	MessagesOut []Message
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

func (c *MessageResult) WithMessage(message Message) {
	c.MessagesOut = append(c.MessagesOut, message)
}

func (c *MessageResult) WithMessages(messages []Message) {
	c.MessagesOut = append(c.MessagesOut, messages...)
}
