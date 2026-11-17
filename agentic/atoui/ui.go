package atoui

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/tmc/a2ui"
	"github.com/vela-ssoc/ssoc-manager/agentic"
)

type UI struct {
	ftb agentic.BaseToolFactory
	log *slog.Logger
}

func NewUI(ftb agentic.BaseToolFactory, log *slog.Logger) *UI {
	return &UI{ftb: ftb, log: log}
}

func (ui *UI) NewEinoTool(cs agentic.ChatSession) ([]tool.BaseTool, error) {
	inst := ui.newInstance(cs)
	return ui.ftb.Builder().
		Add("render_chart", chartCatalogDescription, inst.chart).
		Add("render_metrics", metricsDescription, inst.metrics).
		Add("render_table", tableDescription, inst.table).
		Add("render_detail", detailDescription, inst.detail).
		Add("render_alerts", alertsDescription, inst.alerts).
		Add("render_form", formDescription, inst.form).
		Add("render_confirm", confirmDescription, inst.confirm).
		Add("update_surface", updateSurfaceDescription, inst.updateSurface).
		Add("delete_surface", deleteSurfaceDescription, inst.deleteSurface).
		Build()
}

func (ui *UI) newInstance(cs agentic.ChatSession) *uiInstance {
	return &uiInstance{sess: cs}
}

type uiInstance struct {
	sess agentic.ChatSession
}

type RenderOutput struct {
	SurfaceID string `json:"surface_id,omitempty"`
	Message   string `json:"message,omitempty"`
}

const defaultSuccessMessage = "已渲染到前端画布。"

func (ui *uiInstance) surfaceID(existing, prefix string) (string, error) {
	if existing = strings.TrimSpace(existing); existing != "" {
		if !identifierPattern.MatchString(existing) {
			return "", fmt.Errorf("surface_id %q 不是合法标识符", existing)
		}
		return existing, nil
	}

	return newSurfaceID(prefix)
}

func (ui *uiInstance) snapshot(ctx context.Context, surfaceID string, ops []a2ui.AgentMessage) (*RenderOutput, error) {
	if err := ui.sess.Emit.A2UISnapshot(ctx, surfaceID, ops); err != nil {
		return nil, err
	}

	return &RenderOutput{SurfaceID: surfaceID, Message: defaultSuccessMessage}, nil
}
