package request

import (
	"strings"

	"github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/types"
	"github.com/cloudwego/eino/schema"
)

type AIChatRunAgentInput struct {
	Model          string              `json:"model"    validate:"required"`
	ThreadID       string              `json:"threadId" validate:"required"`
	RunID          string              `json:"runId"    validate:"required"`
	Messages       []types.Message     `json:"messages" validate:"required,gte=1"`
	ParentRunID    *string             `json:"parentRunId,omitempty"`
	State          any                 `json:"state"`
	Tools          []types.Tool        `json:"tools"`
	Context        []types.Context     `json:"context"`
	ForwardedProps any                 `json:"forwardedProps"`
	Resume         []types.ResumeEntry `json:"resume,omitempty"`
}

func (a AIChatRunAgentInput) SchemaMessages(systemPrompt string) []*schema.Message {
	msgs := make([]*schema.Message, 0, len(a.Messages)+1)
	sysp := schema.SystemMessage(systemPrompt)
	msgs = append(msgs, sysp)

	for _, msg := range a.Messages {
		switch msg.Role {
		case types.RoleUser:
			txt := a.parseContent(msg.Content)
			msgs = append(msgs, schema.UserMessage(txt))
		case types.RoleAssistant:
			txt := a.parseContent(msg.Content)
			msgs = append(msgs, schema.AssistantMessage(txt, nil))
		}
	}

	return msgs
}

func (a AIChatRunAgentInput) parseContent(v any) string {
	switch ct := v.(type) {
	case string:
		return ct
	case []any: // types.InputContent
		return a.readInputContent(ct)
	default:
		return ""
	}
}

func (a AIChatRunAgentInput) readInputContent(cts []any) string {
	buf := new(strings.Builder)
	for _, ct := range cts {
		parts, ok := ct.(map[string]any)
		if !ok || parts["type"] != "text" {
			continue
		}
		if msg, _ := parts["text"].(string); msg != "" {
			buf.WriteString(msg)
		}
	}

	return buf.String()
}
