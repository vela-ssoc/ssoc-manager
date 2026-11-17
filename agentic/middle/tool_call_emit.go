package middle

import (
	"context"

	"github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/events"
	"github.com/cloudwego/eino/compose"
	"github.com/vela-ssoc/ssoc-manager/agentic"
)

func NewToolCallEmit(cs agentic.ChatSession) compose.ToolMiddleware {
	mid := &toolCallEmit{cs: cs}

	return compose.ToolMiddleware{
		Invokable: mid.Invokable,
	}
}

type toolCallEmit struct {
	cs agentic.ChatSession
}

func (tce *toolCallEmit) Invokable(next compose.InvokableToolEndpoint) compose.InvokableToolEndpoint {
	return func(ctx context.Context, input *compose.ToolInput) (*compose.ToolOutput, error) {
		callID, name, emit := input.CallID, input.Name, tce.cs.Emit
		startEvent := events.NewToolCallStartEvent(callID, name, events.WithParentMessageID(tce.cs.MessageID))
		if err := emit.Emit(ctx, startEvent); err != nil {
			return nil, err
		}

		if args := input.Arguments; args != "" {
			if err := emit.Emit(ctx, events.NewToolCallArgsEvent(callID, args)); err != nil {
				return nil, err
			}
		}

		result, err := next(ctx, input)
		if err != nil {
			return nil, err
		}
		if err = emit.Emit(ctx, events.NewToolCallEndEvent(callID)); err != nil {
			return nil, err
		}
		if err != nil {
			return result, err
		}

		data := result.Result
		if data == "" {
			data = "{}"
		}
		if err = emit.Emit(ctx, events.NewToolCallResultEvent("tool-"+callID, callID, data)); err != nil {
			return nil, err
		}

		return result, err
	}
}
