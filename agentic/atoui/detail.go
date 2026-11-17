package atoui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/tmc/a2ui"
)

const detailDescription = "把单个对象的字段渲染成键值详情卡，并推送到当前对话的画布。\n\n" +
	"适用：一台探针、一个中继节点、一条告警、一个资产的详细信息；字段较多时用 groups 分组。\n" +
	"不适用：多条同类记录用 render_table；趋势/占比用 render_chart；少量汇总数字用 render_metrics。\n\n" +
	"数据规则：\n" +
	"1. groups 1~8 组，每组 items 1~20 个；label 和 value 都必须是非空字符串。\n" +
	"2. value 是已经格式化好的展示值；缺失值用“—”，不要编造。\n" +
	"3. 字段顺序即展示顺序；相关的字段放在同一个 group 里。\n" +
	"4. 不要把长文本、JSON、SQL 塞进 value；需要展示代码或长文本时改用正文说明。\n\n" +
	"输出行为：详情卡会立即展示；回复正文不要再重复全部字段，可以给一句结论。\n" +
	"成功后返回 surface_id；需要整体替换时带 surface_id 重新调用本工具。"

type DetailItem struct {
	Label string `json:"label" validate:"required" jsonschema:"required" jsonschema_description:"字段名，例如“内网 IP”"`
	Value string `json:"value" validate:"required" jsonschema:"required" jsonschema_description:"展示值，例如“10.0.0.12”；缺失时用“—”"`
}

type DetailGroup struct {
	Title string       `json:"title,omitempty"                          jsonschema_description:"分组标题，例如“基本信息”“网络”"`
	Items []DetailItem `json:"items" validate:"required,min=1,dive"     jsonschema:"required" jsonschema_description:"1~20 个字段"`
}

type RenderDetailInput struct {
	Title     string        `json:"title,omitempty"                          jsonschema_description:"详情卡标题，例如探针主机名"`
	Groups    []DetailGroup `json:"groups" validate:"required,min=1,dive"    jsonschema:"required" jsonschema_description:"1~8 个分组"`
	Summary   string        `json:"summary,omitempty"                        jsonschema_description:"一句话结论，最多 200 字"`
	SurfaceID string        `json:"surface_id,omitempty"                     jsonschema_description:"仅整体替换已有详情卡时传入；一般留空创建新画布"`
}

func (ui *uiInstance) detail(ctx context.Context, in RenderDetailInput) (*RenderOutput, error) {
	if err := ui.sess.Valid(in); err != nil {
		return nil, err
	}

	surfaceID, err := ui.surfaceID(in.SurfaceID, "detail")
	if err != nil {
		return nil, err
	}

	ops, err := buildDetailOps(ctx, surfaceID, in)
	if err != nil {
		return nil, err
	}

	return ui.snapshot(ctx, surfaceID, ops)
}

func buildDetailOps(_ context.Context, surfaceID string, in RenderDetailInput) ([]a2ui.AgentMessage, error) {
	groups, err := validateDetailInput(&in)
	if err != nil {
		return nil, err
	}

	children := make([]string, 0, len(in.Groups)*2+2)
	components := make([]a2ui.Component, 0, len(in.Groups)*8+4)

	if title := strings.TrimSpace(in.Title); title != "" {
		children = append(children, "detail_title")
		components = append(components, text("detail_title", title))
	}

	for gi, group := range in.Groups {
		groupID := fmt.Sprintf("detail_group_%d", gi)
		children = append(children, groupID)

		groupChildren := make([]string, 0, len(group.Items)+1)
		groupComponents := make([]a2ui.Component, 0, len(group.Items)*2+1)

		if title := strings.TrimSpace(group.Title); title != "" {
			titleID := fmt.Sprintf("detail_group_%d_title", gi)
			groupChildren = append(groupChildren, titleID)
			groupComponents = append(groupComponents, caption(titleID, title))
		}

		for ii := range group.Items {
			rowID := fmt.Sprintf("detail_group_%d_row_%d", gi, ii)
			labelID := fmt.Sprintf("detail_group_%d_label_%d", gi, ii)
			valueID := fmt.Sprintf("detail_group_%d_value_%d", gi, ii)

			groupChildren = append(groupChildren, rowID)
			groupComponents = append(groupComponents,
				row(rowID, labelID, valueID),
				weight(caption(labelID, group.Items[ii].Label), 1),
				weight(textBinding(valueID, fmt.Sprintf("/groups/%d/items/%d/value", gi, ii)), 2),
			)
		}

		components = append(components, column(groupID, groupChildren...))
		components = append(components, groupComponents...)

		if gi < len(in.Groups)-1 {
			dividerID := fmt.Sprintf("detail_divider_%d", gi)
			children = append(children, dividerID)
			components = append(components, divider(dividerID))
		}
	}

	if summary := strings.TrimSpace(in.Summary); summary != "" {
		children = append(children, "detail_summary")
		components = append(components, caption("detail_summary", summary))
	}

	components = append([]a2ui.Component{
		card("root", "detail_col"),
		column("detail_col", children...),
	}, components...)

	return createSurface(surfaceID, components, map[string]any{"groups": groups}), nil
}

func validateDetailInput(input *RenderDetailInput) ([]map[string]any, error) {
	if input == nil {
		return nil, errors.New("detail input is nil")
	}

	if n := len(input.Groups); n == 0 || n > 8 {
		return nil, errors.New("groups 数量必须在 1~8 之间")
	}

	groups := make([]map[string]any, 0, len(input.Groups))
	for gi, group := range input.Groups {
		if n := len(group.Items); n == 0 || n > 20 {
			return nil, fmt.Errorf("groups[%d].items 数量必须在 1~20 之间", gi)
		}

		items := make([]map[string]any, 0, len(group.Items))
		for ii, item := range group.Items {
			if strings.TrimSpace(item.Label) == "" {
				return nil, fmt.Errorf("groups[%d].items[%d].label 不能为空", gi, ii)
			}
			if strings.TrimSpace(item.Value) == "" {
				return nil, fmt.Errorf("groups[%d].items[%d].value 不能为空，缺失值请用“—”", gi, ii)
			}

			items = append(items, map[string]any{"value": strings.TrimSpace(item.Value)})
		}

		groups = append(groups, map[string]any{
			"title": strings.TrimSpace(group.Title),
			"items": items,
		})
	}

	return groups, nil
}
