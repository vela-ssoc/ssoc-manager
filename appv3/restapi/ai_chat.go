package restapi

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/NVIDIA/gontainer/v2"
	"github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/events"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/middlewares/skill"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"github.com/vela-ssoc/ssoc-common-mb/eventsource"
	"github.com/vela-ssoc/ssoc-manager/agentic"
	"github.com/vela-ssoc/ssoc-manager/agentic/fskill"
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
	startEvt := events.NewTextMessageStartEvent(msgID)
	if err := emit.Emit(ctx, startEvt); err != nil {
		return err
	}

	sess := agentic.ChatSession{
		ThreadID:  threadID,
		RunID:     runID,
		Model:     input.Model,
		MessageID: msgID,
		Validate:  c.Validator,
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

	homeDir, _ := os.UserHomeDir()
	var skillDir = filepath.Join(homeDir, ".agents/skills")
	skdir := fskill.Dir(skillDir)
	skfs, err := skill.NewBackendFromFilesystem(ctx, &skill.BackendFromFilesystemConfig{
		Backend: skdir,
		BaseDir: skillDir,
	})
	if err != nil {
		ac.log.Error("加载 SKILL 文件系统出错", "err", err)
		_ = emit.Error(ctx, msgID, err)
		return err
	}
	skl, err := skill.NewMiddleware(ctx, &skill.Config{Backend: skfs})
	if err != nil {
		ac.log.Error("加载 SKILL 中间件", "err", err)
		_ = emit.Error(ctx, msgID, err)
		return err
	}

	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "ssoc-ai-chat",
		Description: "SOC 智能运维助手",
		Instruction: agentPersonaPrompt,
		Model:       cm,
		Handlers:    []adk.ChatModelAgentMiddleware{skl},
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: tools,
				ToolCallMiddlewares: []compose.ToolMiddleware{
					middle.NewToolCallEmit(sess),
				},
			},
		},
		GenModelInput: func(_ context.Context, instruction string, input *adk.AgentInput) ([]adk.Message, error) {
			msgs := make([]adk.Message, 0, len(input.Messages)+1)
			if instruction != "" {
				msgs = append(msgs, schema.SystemMessage(instruction))
			}

			return append(msgs, input.Messages...), nil
		},
		MaxIterations: llm.IfMaxStep(20),
	})
	if err != nil {
		ac.log.Error("创建 AI Agent 失败", "err", err)
		_ = emit.Error(ctx, msgID, err)
		return err
	}

	messages := input.SchemaMessages()
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: true,
	})
	iter := runner.Run(ctx, messages)

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			if ctx.Err() == nil {
				ac.log.Error("AI 推理失败", "err", event.Err)
				_ = emit.Error(ctx, msgID, event.Err)
			}
			return nil
		}
		if event.Output == nil || event.Output.MessageOutput == nil {
			continue
		}
		if err = emitAssistantText(ctx, emit, msgID, event.Output.MessageOutput); err != nil {
			if ctx.Err() == nil {
				ac.log.Error("读取 AI 流式输出失败", "err", err)
				_ = emit.Error(ctx, msgID, err)
			}
			return nil
		}
	}

	_ = emit.Emit(ctx, events.NewTextMessageEndEvent(msgID))
	_ = emit.Emit(ctx, events.NewRunFinishedEvent(input.ThreadID, input.RunID))

	return nil
}

// emitAssistantText 把 ADK 的助手消息事件转成 AG-UI 的文本增量事件。
//
// ADK 在流式模式下通过 MessageStream 逐块输出；非流式模式下 Message 是完整消息。
// 工具结果事件会被跳过，但仍然需要关闭其流，避免资源泄漏。
func emitAssistantText(ctx context.Context, emit *agentic.AGUIEmit, msgID string, out *adk.MessageVariant) error {
	if out.Role != schema.Assistant {
		if out.MessageStream != nil {
			out.MessageStream.Close()
		}
		return nil
	}
	if out.IsStreaming {
		if out.MessageStream == nil {
			return nil
		}
		defer out.MessageStream.Close()

		for {
			chunk, err := out.MessageStream.Recv()
			if errors.Is(err, io.EOF) {
				return nil
			}
			if err != nil {
				return err
			}
			if chunk == nil || chunk.Content == "" {
				continue
			}
			if err = emit.Emit(ctx, events.NewTextMessageContentEvent(msgID, chunk.Content)); err != nil {
				return err
			}
		}
	}
	if out.Message != nil && out.Message.Content != "" {
		return emit.Emit(ctx, events.NewTextMessageContentEvent(msgID, out.Message.Content))
	}

	return nil
}
