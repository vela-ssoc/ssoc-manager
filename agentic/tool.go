package agentic

import (
	"context"
	"log/slog"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// EinoToolNewer 为每次运行创建一套独立的 eino 工具。
// 实现方不要在 receiver 上保存 ChatSession，避免跨会话串数据。
type EinoToolNewer interface {
	NewEinoTool(cs ChatSession) ([]tool.BaseTool, error)
}

func NewEinoTools(cs ChatSession, es []EinoToolNewer) ([]tool.BaseTool, error) {
	ts := make([]tool.BaseTool, 0, len(es))
	for _, e := range es {
		tt, err := e.NewEinoTool(cs)
		if err != nil {
			return nil, err
		}
		ts = append(ts, tt...)
	}

	return ts, nil
}

type BaseToolFactory struct {
	valid Validator
	log   *slog.Logger
}

func NewBaseToolFactory(valid Validator, log *slog.Logger) BaseToolFactory {
	return BaseToolFactory{valid: valid, log: log}
}

func (btf BaseToolFactory) Builder() BaseToolBuilder {
	return BaseToolBuilder{valid: btf.valid}
}

type BaseToolBuilder struct {
	valid Validator
	log   *slog.Logger
	tools []tool.BaseTool
	err   error // last error
}

func (btb BaseToolBuilder) Add[T, D any](toolName, toolDesc string, i utils.InvokeFunc[T, D], opts ...utils.Option) BaseToolBuilder {
	f := btb.wrapValidFunc(i)
	if tl, err := utils.InferTool(toolName, toolDesc, f, opts...); err != nil {
		btb.err = err
	} else if tl != nil {
		btb.tools = append(btb.tools, tl)
	}

	return btb
}

func (btb BaseToolBuilder) Build() ([]tool.BaseTool, error) {
	return btb.tools, btb.err
}

func (btb BaseToolBuilder) wrapValidFunc[T, D any](f utils.InvokeFunc[T, D]) utils.InvokeFunc[T, D] {
	if btb.valid == nil {
		return f
	}

	return func(ctx context.Context, in T) (D, error) {
		if err := btb.valid.Validate(in); err != nil {
			var zero D
			btb.log.Warn("参数校验出错", "err", err)
			return zero, err
		}

		return f(ctx, in)
	}
}
