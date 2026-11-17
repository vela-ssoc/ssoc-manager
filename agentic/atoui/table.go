package atoui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/tmc/a2ui"
)

const tableDescription = "把多列明细数据渲染成表格，并推送到当前对话的画布。\n\n" +
	"适用：告警列表、探针清单、连接记录、资产明细等需要逐行对比的数据。\n" +
	"不适用：趋势或占比用 render_chart；1~12 个汇总数字用 render_metrics；单个对象的详情用 render_detail。\n\n" +
	"数据规则：\n" +
	"1. columns 1~12 列，rows 1~200 行；每个 row 必须包含所有 columns.key。\n" +
	"2. 单元格值只能是字符串、数字、布尔或 null；不要嵌套对象或数组。\n" +
	"3. format 由服务端负责格式化：number 千分位、percent 百分号、bytes 字节单位、duration 秒转可读时长、datetime 支持 RFC3339 字符串或 Unix 秒。\n" +
	"4. total 大于实际行数时，页面会提示“共 N 条，当前展示前 M 条”；不要为了凑数编造行。\n\n" +
	"输出行为：表格会立即展示；回复正文不要再贴表格原文，可以给一句结论。\n" +
	"成功后返回 surface_id；需要整体替换时带 surface_id 重新调用本工具。"

type TableColumn struct {
	Key    string `json:"key"              validate:"required" jsonschema:"required"                                                              jsonschema_description:"行对象的字段名，例如“hostname”；只能包含字母、数字和下划线"`
	Label  string `json:"label"            validate:"required" jsonschema:"required"                                                              jsonschema_description:"表头文案"`
	Format string `json:"format,omitempty"                       jsonschema:"enum=text,enum=number,enum=percent,enum=bytes,enum=duration,enum=datetime" jsonschema_description:"展示格式，默认 text"`
}

type RenderTableInput struct {
	Title     string           `json:"title,omitempty"                        jsonschema_description:"表格标题"`
	Columns   []TableColumn    `json:"columns" validate:"required,min=1,dive" jsonschema:"required" jsonschema_description:"1~12 列，顺序即展示顺序"`
	Rows      []map[string]any `json:"rows"    validate:"required,min=1"      jsonschema:"required" jsonschema_description:"1~200 行；键必须与 columns.key 对齐，值只能是字符串、数字、布尔或 null"`
	Total     int              `json:"total,omitempty"                        jsonschema_description:"总记录数；大于 len(rows) 时页面提示已截断"`
	Summary   string           `json:"summary,omitempty"                      jsonschema_description:"一句话结论，最多 200 字"`
	SurfaceID string           `json:"surface_id,omitempty"                   jsonschema_description:"仅整体替换已有表格时传入；一般留空创建新画布"`
}

func (ui *uiInstance) table(ctx context.Context, in RenderTableInput) (*RenderOutput, error) {
	if err := ui.sess.Valid(in); err != nil {
		return nil, err
	}

	surfaceID, err := ui.surfaceID(in.SurfaceID, "table")
	if err != nil {
		return nil, err
	}

	ops, err := buildTableOps(ctx, surfaceID, in)
	if err != nil {
		return nil, err
	}

	return ui.snapshot(ctx, surfaceID, ops)
}

func buildTableOps(_ context.Context, surfaceID string, in RenderTableInput) ([]a2ui.AgentMessage, error) {
	rows, err := validateAndFormatRows(&in)
	if err != nil {
		return nil, err
	}

	children := make([]string, 0, 5)
	components := make([]a2ui.Component, 0, 8+len(in.Columns)*2)

	if title := strings.TrimSpace(in.Title); title != "" {
		children = append(children, "table_title")
		components = append(components, text("table_title", title))
	}

	headerChildren := make([]string, 0, len(in.Columns))
	for i, col := range in.Columns {
		id := fmt.Sprintf("table_head_%d", i)
		headerChildren = append(headerChildren, id)
		components = append(components, weight(caption(id, col.Label), 1))
	}
	children = append(children, "table_header")
	components = append(components, row("table_header", headerChildren...))

	cellIDs := make([]string, 0, len(in.Columns))
	for i, col := range in.Columns {
		id := fmt.Sprintf("table_cell_%d", i)
		cellIDs = append(cellIDs, id)
		components = append(components, weight(textBinding(id, col.Key), 1))
	}

	children = append(children, "table_rows")
	components = append(components,
		listTemplate("table_rows", "/rows", "table_row"),
		row("table_row", cellIDs...),
	)

	total := in.Total
	if total <= 0 {
		total = len(rows)
	}
	if total > len(rows) {
		children = append(children, "table_total")
		components = append(components, caption(
			"table_total",
			fmt.Sprintf("共 %d 条，当前展示前 %d 条", total, len(rows)),
		))
	}

	if summary := strings.TrimSpace(in.Summary); summary != "" {
		children = append(children, "table_summary")
		components = append(components, caption("table_summary", summary))
	}

	components = append([]a2ui.Component{
		card("root", "table_col"),
		column("table_col", children...),
	}, components...)

	return createSurface(surfaceID, components, map[string]any{"rows": rows}), nil
}

func validateAndFormatRows(input *RenderTableInput) ([]map[string]string, error) {
	if input == nil {
		return nil, errors.New("table input is nil")
	}

	if n := len(input.Columns); n == 0 || n > 12 {
		return nil, errors.New("columns 数量必须在 1~12 之间")
	}
	if n := len(input.Rows); n == 0 || n > 200 {
		return nil, errors.New("rows 数量必须在 1~200 之间")
	}

	keys := make(map[string]struct{}, len(input.Columns))
	for i, col := range input.Columns {
		key := strings.TrimSpace(col.Key)
		if !identifierPattern.MatchString(key) {
			return nil, fmt.Errorf("columns[%d].key=%q 不合法，只能包含字母、数字和下划线，且不能以数字开头", i, col.Key)
		}
		if _, exists := keys[key]; exists {
			return nil, fmt.Errorf("columns[%d].key=%q 重复", i, col.Key)
		}
		keys[key] = struct{}{}

		if strings.TrimSpace(col.Label) == "" {
			return nil, fmt.Errorf("columns[%d].label 不能为空", i)
		}

		switch col.Format {
		case "", "text", "number", "percent", "bytes", "duration", "datetime":
		default:
			return nil, fmt.Errorf("columns[%d].format=%q 不支持", i, col.Format)
		}
	}

	rows := make([]map[string]string, 0, len(input.Rows))
	for i, raw := range input.Rows {
		if len(raw) == 0 {
			return nil, fmt.Errorf("rows[%d] 不能为空", i)
		}

		out := make(map[string]string, len(input.Columns))
		for _, col := range input.Columns {
			value, exists := raw[col.Key]
			if !exists {
				return nil, fmt.Errorf("rows[%d] 缺少列 %q", i, col.Key)
			}

			switch value.(type) {
			case map[string]any, []any:
				return nil, fmt.Errorf("rows[%d].%s 不能是对象或数组，请先展开为标量", i, col.Key)
			}

			cell, err := formatCell(value, col.Format)
			if err != nil {
				return nil, fmt.Errorf("rows[%d].%s 格式化失败：%w", i, col.Key, err)
			}
			out[col.Key] = cell
		}
		rows = append(rows, out)
	}

	return rows, nil
}
