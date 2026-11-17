package restapi

import (
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/NVIDIA/gontainer/v2"
	"github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/events"
	"github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/types"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"github.com/vela-ssoc/ssoc-common-mb/eventsource"
	"github.com/vela-ssoc/ssoc-manager/agentic"
	"github.com/vela-ssoc/ssoc-manager/agentic/middle"
	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"github.com/vela-ssoc/ssoc-manager/appv3/service"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/xgfone/ship/v5"
)

type AIChat struct {
	pdb     polydb.Sourcer
	svc     *service.AIChat
	tools   []agentic.EinoToolNewer
	httpcli *http.Client
	log     *slog.Logger
}

func NewAIChat(pdb polydb.Sourcer, svc *service.AIChat, tools gontainer.Multiple[agentic.EinoToolNewer], httpcli *http.Client, log *slog.Logger) *AIChat {
	return &AIChat{
		pdb:     pdb,
		svc:     svc,
		tools:   tools,
		httpcli: httpcli,
		log:     log,
	}
}

func (ac *AIChat) RegisterHTTP(g shipx.RouteGroup) error {
	g.ExposeV3.Group.Route("/sse/ai-chat/run").POST(ac.run)

	return nil
}

func (ac *AIChat) run(c *ship.Context) error {
	input := new(request.AIChatRunAgentInput)
	if err := c.Bind(input); err != nil {
		return err
	}

	w, r := c.Response(), c.Request()
	ctx := r.Context()
	sse := eventsource.Accept(w, r)
	if sse == nil {
		return ship.ErrUnsupportedMediaType
	}
	defer sse.Close()

	emit := agentic.NewAGUIEmit(sse)
	threadID, runID := input.ThreadID, input.RunID
	if err := emit.Emit(ctx, events.NewRunStartedEvent(threadID, runID)); err != nil {
		ac.log.Error("消息发送出错", "err", err)
		return err
	}

	msgID := events.GenerateMessageID()
	startEvt := events.NewTextMessageStartEvent(msgID, events.WithRole(string(types.RoleAssistant)))
	if err := emit.Emit(ctx, startEvt); err != nil {
		return err
	}

	sess := agentic.ChatSession{
		ThreadID:  threadID,
		RunID:     runID,
		Model:     input.Model,
		MessageID: msgID,
		Validate:  c.Validator.Validate,
		Emit:      emit,
	}
	llm, err := ac.svc.LLM(ctx, input.Model)
	if err != nil {
		_ = emit.Error(ctx, msgID, err)
		ac.log.Error("加载大模型出错", "err", err)
		return err
	}

	cfg := &openai.ChatModelConfig{
		APIKey:     llm.APIKey,
		BaseURL:    llm.BaseURL,
		Model:      input.Model,
		HTTPClient: ac.httpcli,
	}
	cm, err := openai.NewChatModel(ctx, cfg)
	if err != nil {
		ac.log.Error("初始化 AI 模型失败", "err", err)
		_ = emit.Error(ctx, msgID, err)
		return err
	}

	tools, err := agentic.NewEinoTools(sess, ac.tools)
	if err != nil {
		ac.log.Error("创建 Agent Tool 出错", "err", err)
		_ = emit.Error(ctx, msgID, err)
		return err
	}

	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: cm,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: tools,
			ToolCallMiddlewares: []compose.ToolMiddleware{
				middle.NewToolCallEmit(sess),
			},
		},
		MaxStep: llm.IfMaxStep(20),
	})
	if err != nil {
		ac.log.Error("创建 AI Agent 失败", "err", err)
		_ = emit.Error(ctx, msgID, err)
		return nil
	}

	messages := input.SchemaMessages(llm.Prompt)
	stream, err := agent.Stream(ctx, messages)
	if err != nil {
		ac.log.Error("AI 推理失败", "err", err)
		_ = emit.Error(ctx, msgID, err)
		return nil
	}
	defer stream.Close()

	for {
		chunk, err1 := stream.Recv()
		if errors.Is(err1, io.EOF) {
			break
		}
		if err1 != nil {
			if ctx.Err() == nil {
				ac.log.Error("读取 AI 流式输出失败", "err", err1)
				_ = emit.Error(ctx, msgID, err1)
			}
			return nil
		}
		if chunk == nil || chunk.Content == "" {
			continue
		}

		if err1 = emit.Emit(ctx, events.NewTextMessageContentEvent(msgID, chunk.Content)); err1 != nil {
			return nil
		}
	}

	_ = emit.Emit(ctx, events.NewTextMessageEndEvent(msgID))
	_ = emit.Emit(ctx, events.NewRunFinishedEvent(input.ThreadID, input.RunID))

	return nil
}
