package atoui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/tmc/a2ui"
)

const chartCatalogDescription = "把真实业务数据渲染成图表，并推送到当前对话的画布。\n\n" +
	"适用：时间趋势（line）、分类对比（bar）、整体占比（pie）。\n" +
	"不适用：单个或少量指标用 render_metrics；多列明细用 render_table；需要用户输入用 render_form。\n\n" +
	"数据规则：\n" +
	"1. 只能使用工具查询到的真实数据，禁止编造、估算或补全缺失值。\n" +
	"2. x_axis 与每个 series.data 的长度必须完全一致；line/bar 最多 100 个分类点、10 个系列；pie 只能 1 个系列，数值非负且总和大于 0。\n" +
	"3. 数值必须是有限数；百分比传 0-100 的数值，不要传“12%”这样的字符串。\n" +
	"4. 标题要说明统计对象和时间口径，例如“近 7 天在线探针数（按天）”。\n\n" +
	"输出行为：\n" +
	"图表会立即在画布展示，回复正文不要再重复数据表、坐标点或 A2UI JSON；可以给一句结论。\n" +
	"成功后返回 surface_id；需要整体替换这张图时，带 surface_id 重新调用本工具。"

type ChartSeries struct {
	Name string    `json:"name" validate:"required" jsonschema:"required" jsonschema_description:"系列名称，作为图例，例如“在线”“离线”"`
	Data []float64 `json:"data" validate:"required,min=1" jsonschema:"required" jsonschema_description:"数值数组，长度必须与 x_axis 完全一致；百分比传 0-100 的数值"`
}

type RenderChartInput struct {
	Title     string        `json:"title"                 validate:"required"                              jsonschema:"required"                            jsonschema_description:"图表标题，说明统计对象与口径，例如“近 7 天在线探针数（按天）”"`
	ChartType string        `json:"chart_type"            validate:"required,oneof=line bar pie"            jsonschema:"required,enum=line,enum=bar,enum=pie" jsonschema_description:"line=趋势，bar=分类对比，pie=占比"`
	XAxis     []string      `json:"x_axis"                validate:"required,min=1,dive,required"           jsonschema:"required"                            jsonschema_description:"分类标签，1~100 个；饼图为扇区名称"`
	Series    []ChartSeries `json:"series"                validate:"required,min=1,dive"                     jsonschema:"required"                            jsonschema_description:"数据系列，1~10 个；饼图只能 1 个"`
	Summary   string        `json:"summary,omitempty"                                                      jsonschema_description:"一句话结论，最多 200 字"`
	SurfaceID string        `json:"surface_id,omitempty"                                                   jsonschema_description:"仅整体替换已有图表时传入；一般留空创建新画布"`
}

func (ui *uiInstance) chart(ctx context.Context, in RenderChartInput) (*RenderOutput, error) {
	if err := ui.sess.Valid(in); err != nil {
		return nil, err
	}

	surfaceID, err := ui.surfaceID(in.SurfaceID, "chart")
	if err != nil {
		return nil, err
	}

	ops, err := buildChartOps(ctx, surfaceID, in)
	if err != nil {
		return nil, err
	}

	return ui.snapshot(ctx, surfaceID, ops)
}

func buildChartOps(_ context.Context, surfaceID string, in RenderChartInput) ([]a2ui.AgentMessage, error) {
	if err := validateChartInput(&in); err != nil {
		return nil, err
	}

	properties := make(map[string]json.RawMessage, 5)

	values := map[string]any{
		"title":      in.Title,
		"chart_type": in.ChartType,
		"x_axis":     in.XAxis,
		"series":     in.Series,
	}
	if summary := strings.TrimSpace(in.Summary); summary != "" {
		values["summary"] = summary
	}

	for key, value := range values {
		raw, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("marshal chart property %q: %w", key, err)
		}
		properties[key] = raw
	}

	root := a2ui.Component{
		ID: "root",
		Custom: &a2ui.CustomComponent{
			Type:       ChartComponent,
			Properties: properties,
		},
	}

	return createSurface(surfaceID, []a2ui.Component{root}, nil), nil
}

func validateChartInput(input *RenderChartInput) error {
	if input == nil {
		return errors.New("chart input is nil")
	}

	if strings.TrimSpace(input.Title) == "" {
		return errors.New("title 不能为空")
	}

	switch input.ChartType {
	case "line", "bar", "pie":
	default:
		return fmt.Errorf("chart_type=%q 不支持，只能是 line、bar、pie", input.ChartType)
	}

	if n := len(input.XAxis); n == 0 || n > 100 {
		return errors.New("x_axis 长度必须在 1~100 之间")
	}

	if n := len(input.Series); n == 0 || n > 10 {
		return errors.New("series 数量必须在 1~10 之间")
	}

	for i, label := range input.XAxis {
		if strings.TrimSpace(label) == "" {
			return fmt.Errorf("x_axis[%d] 不能为空", i)
		}
	}

	names := make(map[string]struct{}, len(input.Series))

	for i, series := range input.Series {
		name := strings.TrimSpace(series.Name)
		if name == "" {
			return fmt.Errorf("series[%d].name 不能为空", i)
		}

		if _, exists := names[name]; exists {
			return fmt.Errorf("series 名称 %q 重复，请为每个系列使用不同名称", name)
		}
		names[name] = struct{}{}

		if len(series.Data) != len(input.XAxis) {
			return fmt.Errorf(
				"series[%d].data 长度 %d 与 x_axis 长度 %d 不一致",
				i, len(series.Data), len(input.XAxis),
			)
		}

		for j, value := range series.Data {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return fmt.Errorf("series[%d].data[%d] 不是有限数", i, j)
			}

			if input.ChartType == "pie" && value < 0 {
				return fmt.Errorf("饼图的 series[%d].data[%d] 不能为负数", i, j)
			}
		}
	}

	if input.ChartType == "pie" {
		if len(input.Series) != 1 {
			return errors.New("饼图只能提供 1 个 series")
		}

		var total float64
		for _, value := range input.Series[0].Data {
			total += value
		}
		if total <= 0 {
			return errors.New("饼图数据总和必须大于 0")
		}
	}

	return nil
}
