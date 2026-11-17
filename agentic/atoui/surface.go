package atoui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/tmc/a2ui"
)

const updateSurfaceDescription = "局部更新一个已经渲染的 A2UI 画布，使用 ACTIVITY_DELTA 追加数据模型变更。\n\n" +
	"适用：render_metrics、render_table、render_detail、render_alerts、render_form 创建后，只需要换数据、不改变布局时。\n" +
	"不适用：图表（Chart）的数据绑定是组件属性，请带 surface_id 重新调用 render_chart；需要改变布局时也请重新调用对应的 render_* 工具。\n\n" +
	"数据规则：\n" +
	"1. surface_id 必须来自当前对话中 render_* 的返回值；不要编造。\n" +
	"2. mode=merge（默认）按顶层 key 合并，key 传 null 表示删除该 key；mode=replace 会用 data 整体替换数据模型。\n" +
	"3. data 的 key 必须与创建 surface 时使用的数据模型 key 一致，例如 metrics 用 items、table 用 rows、detail 用 groups、alerts 用 items、form 用 form。\n" +
	"4. 只发变化的部分，不要重复发送未变化的大字段。\n\n" +
	"输出行为：画布会原地更新，不会新增消息；回复正文不要再重复数据。"

const deleteSurfaceDescription = "移除一个已经渲染的 A2UI 画布。\n\n" +
	"适用：内容已经过期、用户明确要求关闭、或同一内容已被新的 surface 替换时。\n" +
	"不适用：只想换数据时用 update_surface；不要用它清理不存在的 surface。\n" +
	"surface_id 必须来自当前对话中 render_* 的返回值。"

type UpdateSurfaceInput struct {
	SurfaceID string         `json:"surface_id" validate:"required"        jsonschema:"required"                    jsonschema_description:"render_* 返回的画布 ID"`
	Data      map[string]any `json:"data"       validate:"required,min=1"   jsonschema:"required"                    jsonschema_description:"写入数据模型的值；mode=merge 时按顶层 key 合并，key 传 null 表示删除"`
	Mode      string         `json:"mode,omitempty" validate:"omitempty,oneof=merge replace" jsonschema:"enum=merge,enum=replace" jsonschema_description:"merge=按 key 合并（默认），replace=整体替换数据模型"`
}

type DeleteSurfaceInput struct {
	SurfaceID string `json:"surface_id" validate:"required" jsonschema:"required" jsonschema_description:"要移除的画布 ID"`
}

func (ui *uiInstance) updateSurface(ctx context.Context, in UpdateSurfaceInput) (*RenderOutput, error) {
	if err := ui.sess.Valid(in); err != nil {
		return nil, err
	}

	ops, err := updateSurfaceOps(&in)
	if err != nil {
		return nil, err
	}

	if err = ui.sess.Emit.A2UIDelta(ctx, in.SurfaceID, ops); err != nil {
		return nil, err
	}

	return &RenderOutput{
		SurfaceID: in.SurfaceID,
		Message:   "已更新画布。请用一句话说明变化，不要重复数据。",
	}, nil
}

func updateSurfaceOps(input *UpdateSurfaceInput) ([]a2ui.AgentMessage, error) {
	if input == nil {
		return nil, errors.New("update surface input is nil")
	}

	input.SurfaceID = strings.TrimSpace(input.SurfaceID)
	if !identifierPattern.MatchString(input.SurfaceID) {
		return nil, errors.New("surface_id 不合法")
	}
	if len(input.Data) == 0 {
		return nil, errors.New("data 不能为空")
	}

	switch input.Mode {
	case "", "merge":
		ops := make([]a2ui.AgentMessage, 0, len(input.Data))
		for key, value := range input.Data {
			if strings.TrimSpace(key) == "" {
				return nil, errors.New("data 的 key 不能为空")
			}
			ops = append(ops, a2ui.AgentMessage{
				Version: a2ui.Version,
				UpdateDataModel: &a2ui.UpdateDataModel{
					SurfaceID: input.SurfaceID,
					Path:      pointerJoin("/", key),
					Value:     value,
				},
			})
		}
		return ops, nil
	case "replace":
		return []a2ui.AgentMessage{{
			Version: a2ui.Version,
			UpdateDataModel: &a2ui.UpdateDataModel{
				SurfaceID: input.SurfaceID,
				Path:      "/",
				Value:     input.Data,
			},
		}}, nil
	default:
		return nil, fmt.Errorf("mode=%q 不支持，只能是 merge、replace 或留空", input.Mode)
	}
}

func (ui *uiInstance) deleteSurface(ctx context.Context, in DeleteSurfaceInput) (*RenderOutput, error) {
	if err := ui.sess.Valid(in); err != nil {
		return nil, err
	}

	surfaceID := strings.TrimSpace(in.SurfaceID)
	if !identifierPattern.MatchString(surfaceID) {
		return nil, errors.New("surface_id 不合法")
	}

	if err := ui.sess.Emit.A2UIDelete(ctx, surfaceID); err != nil {
		return nil, err
	}

	return &RenderOutput{
		SurfaceID: surfaceID,
		Message:   "已移除画布。",
	}, nil
}
