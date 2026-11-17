package atoui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/tmc/a2ui"
)

const confirmDescription = "渲染一个危险操作确认卡，等待用户点击确认或取消。\n\n" +
	"适用：删除、结束连接、重启、批量变更等不可逆或影响范围较大的操作；执行前必须让用户明确确认。\n" +
	"不适用：普通参数收集用 render_form；只读展示用 render_detail / render_alerts。\n\n" +
	"数据规则：\n" +
	"1. message 必须写清影响对象、影响范围和不可逆性，不要只写“是否继续”。\n" +
	"2. tone=danger 用于不可逆操作，warning 用于可恢复但影响较大的操作。\n" +
	"3. context 只支持字符串、数字、布尔或由它们组成的数组；确认结果会以 confirmed=true/false 回传。\n" +
	"4. 用户点击后会以 action_name 开启新一轮请求；不要在确认前直接执行操作。\n\n" +
	"输出行为：确认卡会立即展示；回复正文不要再重复风险说明。\n" +
	"成功后返回 surface_id；需要整体替换时带 surface_id 重新调用本工具。"

type RenderConfirmInput struct {
	Title        string         `json:"title"                   validate:"required"                     jsonschema:"required"                 jsonschema_description:"确认框标题，例如“结束探针子流”"`
	Message      string         `json:"message"                 validate:"required"                     jsonschema:"required"                 jsonschema_description:"风险说明，必须写清影响对象、影响范围和不可逆性"`
	Tone         string         `json:"tone,omitempty"          validate:"omitempty,oneof=warning danger" jsonschema:"enum=warning,enum=danger" jsonschema_description:"danger 用于不可逆操作；默认 warning"`
	ConfirmLabel string         `json:"confirm_label,omitempty"                                          jsonschema_description:"确认按钮文案，默认“确认”"`
	CancelLabel  string         `json:"cancel_label,omitempty"                                           jsonschema_description:"取消按钮文案，默认“取消”"`
	ActionName   string         `json:"action_name"             validate:"required"                     jsonschema:"required"                 jsonschema_description:"确认后回传的事件名，例如 stop_substream"`
	Context      map[string]any `json:"context,omitempty"                                                jsonschema_description:"确认后回传的上下文，例如 id、slug"`
	SurfaceID    string         `json:"surface_id,omitempty"                                             jsonschema_description:"仅整体替换已有确认卡时传入；一般留空创建新画布"`
}

func (ui *uiInstance) confirm(ctx context.Context, in RenderConfirmInput) (*RenderOutput, error) {
	if err := ui.sess.Valid(in); err != nil {
		return nil, err
	}

	surfaceID, err := ui.surfaceID(in.SurfaceID, "confirm")
	if err != nil {
		return nil, err
	}

	ops, err := buildConfirmOps(ctx, surfaceID, in)
	if err != nil {
		return nil, err
	}

	return ui.snapshot(ctx, surfaceID, ops)
}

func buildConfirmOps(_ context.Context, surfaceID string, in RenderConfirmInput) ([]a2ui.AgentMessage, error) {
	if err := validateConfirmInput(&in); err != nil {
		return nil, err
	}

	tone := strings.ToLower(strings.TrimSpace(in.Tone))
	iconName := a2ui.IconWarning
	if tone == "danger" {
		iconName = a2ui.IconError
	}

	confirmLabel := strings.TrimSpace(in.ConfirmLabel)
	if confirmLabel == "" {
		confirmLabel = "确认"
	}
	cancelLabel := strings.TrimSpace(in.CancelLabel)
	if cancelLabel == "" {
		cancelLabel = "取消"
	}

	baseCtx := dynamicValues(in.Context)
	if baseCtx == nil {
		baseCtx = make(map[string]a2ui.DynamicValue, 2)
	}

	confirmCtx := make(map[string]a2ui.DynamicValue, len(baseCtx)+1)
	cancelCtx := make(map[string]a2ui.DynamicValue, len(baseCtx)+1)
	for key, value := range baseCtx {
		confirmCtx[key] = value
		cancelCtx[key] = value
	}
	confirmCtx["confirmed"] = a2ui.ValueBool(true)
	cancelCtx["confirmed"] = a2ui.ValueBool(false)

	components := []a2ui.Component{
		card("root", "confirm_col"),
		column("confirm_col", "confirm_head", "confirm_message", "confirm_actions"),
		row("confirm_head", "confirm_icon", "confirm_title"),
		icon("confirm_icon", iconName),
		text("confirm_title", strings.TrimSpace(in.Title)),
		text("confirm_message", strings.TrimSpace(in.Message)),
		row("confirm_actions", "confirm_ok", "confirm_cancel"),
		button("confirm_ok", "confirm_ok_label", eventAction(in.ActionName, confirmCtx), a2ui.ButtonVariantPrimary),
		text("confirm_ok_label", confirmLabel),
		button("confirm_cancel", "confirm_cancel_label", eventAction(in.ActionName, cancelCtx), a2ui.ButtonVariantDefault),
		text("confirm_cancel_label", cancelLabel),
	}

	return createSurface(surfaceID, components, nil), nil
}

func validateConfirmInput(input *RenderConfirmInput) error {
	if input == nil {
		return errors.New("confirm input is nil")
	}

	if strings.TrimSpace(input.Title) == "" {
		return errors.New("title 不能为空")
	}
	if strings.TrimSpace(input.Message) == "" {
		return errors.New("message 不能为空")
	}

	input.ActionName = strings.TrimSpace(input.ActionName)
	if input.ActionName == "" {
		return errors.New("action_name 不能为空")
	}
	if !identifierPattern.MatchString(input.ActionName) {
		return errors.New("action_name 只能包含字母、数字和下划线，且不能以数字开头")
	}

	switch strings.ToLower(strings.TrimSpace(input.Tone)) {
	case "", "warning", "danger":
	default:
		return fmt.Errorf("tone=%q 不支持，只能是 warning、danger 或留空", input.Tone)
	}

	return validateActionContext(input.Context)
}
