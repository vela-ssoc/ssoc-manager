package agentic

// Validator is used to validate the v is valid or not.
type Validator interface {
	Validate(v any) error
}

type ChatSession struct {
	ThreadID  string
	RunID     string
	Model     string
	MessageID string
	Validate  Validator
	Emit      *AGUIEmit
}

func (cs ChatSession) Valid(arg any) error {
	if v := cs.Validate; v != nil {
		return v.Validate(arg)
	}

	return nil
}
