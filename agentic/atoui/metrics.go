package atoui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/tmc/a2ui"
)

const metricsDescription = "把 1~12 个核心指标渲染成一组指标卡，并推送到当前对话的画布。\n\n" +
	"适用：在线数、告警数、覆盖率、成功率等需要一眼看出的汇总数字；需要对比变化时带上 delta。\n" +
	"不适用：趋势、占比用 render_chart；多列明细用 render_table；单个对象的多个字段用 render_detail。\n\n" +
	"数据规则：\n" +
	"1. 只能使用真实数据；value 是已经格式化好的展示值，不要在这里写计算公式或 SQL。\n" +
	"2. value 不要带单位，单位单独放在 unit；百分比数值传 0-100。\n" +
	"3. delta 是可选的环比/同比变化，例如“+12.5%”“-3”；没有对比数据时留空，不要编造。\n" +
	"4. 指标数量控制在 1~12 个，优先展示结论性的核心指标。\n\n" +
	"输出行为：指标卡会立即展示；回复正文不要再重复这些数字，可以给一句结论。\n" +
	"成功后返回 surface_id；需要整体替换时带 surface_id 重新调用本工具。"

type MetricItem struct {
	Label string `json:"label"           validate:"required" jsonschema:"required"                                           jsonschema_description:"指标名称，例如“在线探针”"`
	Value string `json:"value"           validate:"required" jsonschema:"required"                                           jsonschema_description:"格式化后的值，例如“1,284”“99.2”；不要带单位"`
	Unit  string `json:"unit,omitempty"                                                                                      jsonschema_description:"单位，例如“台”“%”“ms”"`
	Delta string `json:"delta,omitempty"                                                                                     jsonschema_description:"环比变化，例如“+12.5%”“-3”；没有对比数据时留空"`
	Trend string `json:"trend,omitempty"                     jsonschema:"enum=up,enum=down,enum=flat"                        jsonschema_description:"变化方向，用于在 delta 前加箭头"`
}

type RenderMetricsInput struct {
	Title     string       `json:"title,omitempty"                        jsonschema_description:"指标组标题"`
	Items     []MetricItem `json:"items"   validate:"required,min=1,dive" jsonschema:"required" jsonschema_description:"1~12 个指标；每个指标必须有 label 和 value"`
	Summary   string       `json:"summary,omitempty"                      jsonschema_description:"一句话结论，最多 200 字"`
	SurfaceID string       `json:"surface_id,omitempty"                   jsonschema_description:"仅整体替换已有指标组时传入；一般留空创建新画布"`
}

func (ui *uiInstance) metrics(ctx context.Context, in RenderMetricsInput) (*RenderOutput, error) {
	if err := ui.sess.Valid(in); err != nil {
		return nil, err
	}

	surfaceID, err := ui.surfaceID(in.SurfaceID, "metrics")
	if err != nil {
		return nil, err
	}

	ops, err := buildMetricsOps(ctx, surfaceID, in)
	if err != nil {
		return nil, err
	}

	return ui.snapshot(ctx, surfaceID, ops)
}

func buildMetricsOps(_ context.Context, surfaceID string, in RenderMetricsInput) ([]a2ui.AgentMessage, error) {
	if err := validateMetricsInput(&in); err != nil {
		return nil, err
	}

	items := make([]map[string]any, 0, len(in.Items))
	for _, item := range in.Items {
		items = append(items, map[string]any{
			"label": strings.TrimSpace(item.Label),
			"value": joinValueUnit(item.Value, item.Unit),
			"delta": trendText(item.Trend, item.Delta),
		})
	}

	children := make([]string, 0, 3)
	components := make([]a2ui.Component, 0, 8)

	if title := strings.TrimSpace(in.Title); title != "" {
		children = append(children, "metrics_title")
		components = append(components, text("metrics_title", title))
	}

	children = append(children, "metrics_items")
	components = append(components,
		listTemplate("metrics_items", "/items", "metrics_row"),
		row("metrics_row", "metrics_label", "metrics_value", "metrics_delta"),
		weight(captionBinding("metrics_label", "label"), 2),
		weight(textBinding("metrics_value", "value"), 1),
		weight(captionBinding("metrics_delta", "delta"), 1),
	)

	if summary := strings.TrimSpace(in.Summary); summary != "" {
		children = append(children, "metrics_summary")
		components = append(components, caption("metrics_summary", summary))
	}

	components = append([]a2ui.Component{
		card("root", "metrics_col"),
		column("metrics_col", children...),
	}, components...)

	return createSurface(surfaceID, components, map[string]any{"items": items}), nil
}

func validateMetricsInput(input *RenderMetricsInput) error {
	if input == nil {
		return errors.New("metrics input is nil")
	}

	if n := len(input.Items); n == 0 || n > 12 {
		return errors.New("items 数量必须在 1~12 之间")
	}

	for i, item := range input.Items {
		if strings.TrimSpace(item.Label) == "" {
			return fmt.Errorf("items[%d].label 不能为空", i)
		}
		if strings.TrimSpace(item.Value) == "" {
			return fmt.Errorf("items[%d].value 不能为空", i)
		}

		switch item.Trend {
		case "", "up", "down", "flat":
		default:
			return fmt.Errorf("items[%d].trend=%q 不支持，只能是 up、down、flat 或留空", i, item.Trend)
		}
	}

	return nil
}
