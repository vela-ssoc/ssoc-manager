package agentic

type ChatSession struct {
	ThreadID  string
	RunID     string
	Model     string
	MessageID string
	Validate  func(any) error
	Emit      *AGUIEmit
}

func (cs ChatSession) Valid(v any) error {
	if f := cs.Validate; f != nil {
		return f(v)
	}

	return nil
}
