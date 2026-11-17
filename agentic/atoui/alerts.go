package atoui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/tmc/a2ui"
)

const alertsDescription = "把告警、事件或操作记录渲染成带状态图标的列表，并推送到当前对话的画布。\n\n" +
	"适用：最近告警、探针上下线、任务执行记录、审计事件等按时间排列的条目。\n" +
	"不适用：需要逐列对比的明细用 render_table；单个对象详情用 render_detail；趋势/占比用 render_chart。\n\n" +
	"数据规则：\n" +
	"1. items 1~50 条，按时间倒序排列；title 必须非空。\n" +
	"2. level 只能是 info、success、warning、danger，用来选择状态图标；不要用 level 表达业务分类。\n" +
	"3. time 是已经格式化好的展示时间，例如“2026-11-13 10:21:03”；没有时间时留空。\n" +
	"4. description 是可选的补充说明，保持一句话；不要贴大段日志。\n\n" +
	"输出行为：列表会立即展示；回复正文不要再重复条目，可以给一句结论。\n" +
	"成功后返回 surface_id；需要整体替换时带 surface_id 重新调用本工具。"

type AlertItem struct {
	Time        string `json:"time,omitempty"                jsonschema_description:"展示时间，例如“2026-11-13 10:21:03”"`
	Level       string `json:"level,omitempty"               jsonschema:"enum=info,enum=success,enum=warning,enum=danger" jsonschema_description:"状态级别，决定图标；默认 info"`
	Title       string `json:"title"  validate:"required"    jsonschema:"required"  jsonschema_description:"条目主文案，例如“探针 node-01 已上线”"`
	Description string `json:"description,omitempty"         jsonschema_description:"补充说明，一句话"`
}

type RenderAlertsInput struct {
	Title     string      `json:"title,omitempty"                          jsonschema_description:"列表标题，例如“最近告警”"`
	Items     []AlertItem `json:"items" validate:"required,min=1,dive"     jsonschema:"required" jsonschema_description:"1~50 条，按时间倒序"`
	Summary   string      `json:"summary,omitempty"                        jsonschema_description:"一句话结论，最多 200 字"`
	SurfaceID string      `json:"surface_id,omitempty"                     jsonschema_description:"仅整体替换已有列表时传入；一般留空创建新画布"`
}

func (ui *uiInstance) alerts(ctx context.Context, in RenderAlertsInput) (*RenderOutput, error) {
	if err := ui.sess.Valid(in); err != nil {
		return nil, err
	}

	surfaceID, err := ui.surfaceID(in.SurfaceID, "alerts")
	if err != nil {
		return nil, err
	}

	ops, err := buildAlertsOps(ctx, surfaceID, in)
	if err != nil {
		return nil, err
	}

	return ui.snapshot(ctx, surfaceID, ops)
}

func buildAlertsOps(_ context.Context, surfaceID string, in RenderAlertsInput) ([]a2ui.AgentMessage, error) {
	items, err := validateAlertsInput(&in)
	if err != nil {
		return nil, err
	}

	children := make([]string, 0, 3)
	components := make([]a2ui.Component, 0, 8)

	if title := strings.TrimSpace(in.Title); title != "" {
		children = append(children, "alerts_title")
		components = append(components, text("alerts_title", title))
	}

	children = append(children, "alerts_items")
	components = append(components,
		listTemplate("alerts_items", "/items", "alerts_item"),
		row("alerts_item", "alerts_icon", "alerts_body"),
		iconBinding("alerts_icon", "icon"),
		column("alerts_body", "alerts_item_title", "alerts_item_time", "alerts_item_desc"),
		textBinding("alerts_item_title", "title"),
		captionBinding("alerts_item_time", "time"),
		captionBinding("alerts_item_desc", "desc"),
	)

	if summary := strings.TrimSpace(in.Summary); summary != "" {
		children = append(children, "alerts_summary")
		components = append(components, caption("alerts_summary", summary))
	}

	components = append([]a2ui.Component{
		card("root", "alerts_col"),
		column("alerts_col", children...),
	}, components...)

	return createSurface(surfaceID, components, map[string]any{"items": items}), nil
}

func validateAlertsInput(input *RenderAlertsInput) ([]map[string]any, error) {
	if input == nil {
		return nil, errors.New("alerts input is nil")
	}

	if n := len(input.Items); n == 0 || n > 50 {
		return nil, errors.New("items 数量必须在 1~50 之间")
	}

	items := make([]map[string]any, 0, len(input.Items))
	for i, item := range input.Items {
		if strings.TrimSpace(item.Title) == "" {
			return nil, fmt.Errorf("items[%d].title 不能为空", i)
		}

		iconName, err := alertIcon(item.Level)
		if err != nil {
			return nil, fmt.Errorf("items[%d].%w", i, err)
		}

		items = append(items, map[string]any{
			"icon":  iconName,
			"title": strings.TrimSpace(item.Title),
			"time":  strings.TrimSpace(item.Time),
			"desc":  strings.TrimSpace(item.Description),
		})
	}

	return items, nil
}

func alertIcon(level string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "", "info":
		return string(a2ui.IconInfo), nil
	case "success":
		return string(a2ui.IconCheck), nil
	case "warning":
		return string(a2ui.IconWarning), nil
	case "danger":
		return string(a2ui.IconError), nil
	default:
		return "", fmt.Errorf("level=%q 不支持，只能是 info、success、warning、danger", level)
	}
}
