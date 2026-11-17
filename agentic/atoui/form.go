package atoui

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/tmc/a2ui"
)

const formDescription = "把一组输入控件渲染成表单，收集用户参数后回传给 Agent。\n\n" +
	"适用：需要用户确认筛选条件、补充查询参数、选择时间范围、填写执行参数时。\n" +
	"不适用：只读展示用 render_detail / render_table；危险操作确认用 render_confirm；不要为了展示数据调用本工具。\n\n" +
	"数据规则：\n" +
	"1. fields 1~20 个；name 只能包含字母、数字和下划线，且不能以数字开头；name 不能重复。\n" +
	"2. select / multiselect 必须提供 options；slider 必须满足 max > min，step 可选且必须大于 0。\n" +
	"3. default 的类型必须与控件匹配：text 用字符串，number/slider 用数字，checkbox 用布尔，multiselect 用字符串数组。\n" +
	"4. context 只支持字符串、数字、布尔或由它们组成的数组；字段值会自动随提交回传，不需要在 context 里重复。\n" +
	"5. 用户提交后会以 action_name 开启新一轮请求，字段值放在 context 中以字段名作为 key。\n\n" +
	"输出行为：表单会立即展示；回复正文不要再重复字段定义。\n" +
	"成功后返回 surface_id；需要整体替换时带 surface_id 重新调用本工具。"

type FormOption struct {
	Label string `json:"label" validate:"required" jsonschema:"required" jsonschema_description:"选项文案"`
	Value string `json:"value" validate:"required" jsonschema:"required" jsonschema_description:"选项值，提交时回传"`
}

type FormField struct {
	Name        string       `json:"name"                  validate:"required"                                                                                                                         jsonschema:"required"                                                                                                                     jsonschema_description:"数据模型键，提交时回传；只能包含字母、数字和下划线"`
	Label       string       `json:"label"                 validate:"required"                                                                                                                         jsonschema:"required"                                                                                                                     jsonschema_description:"字段标签"`
	Type        string       `json:"type"                  validate:"required"                                                                                                                         jsonschema:"required,enum=text,enum=textarea,enum=number,enum=select,enum=multiselect,enum=checkbox,enum=slider,enum=date,enum=datetime" jsonschema_description:"控件类型"`
	Required    bool         `json:"required,omitempty"                                                                                                                                                   jsonschema_description:"是否必填；必填字段会附加 required 校验"`
	Placeholder string       `json:"placeholder,omitempty"                                                                                                                                                jsonschema_description:"输入提示"`
	Options     []FormOption `json:"options,omitempty"                                                                                                                                                    jsonschema_description:"type=select/multiselect 时必填"`
	Default     any          `json:"default,omitempty"                                                                                                                                                    jsonschema_description:"默认值，类型必须与控件匹配"`
	Min         *float64     `json:"min,omitempty"                                                                                                                                                        jsonschema_description:"number/slider 的最小值"`
	Max         *float64     `json:"max,omitempty"                                                                                                                                                        jsonschema_description:"number/slider 的最大值"`
	Step        *float64     `json:"step,omitempty"                                                                                                                                                       jsonschema_description:"number/slider 的步长，必须大于 0"`
}

type RenderFormInput struct {
	Title       string         `json:"title,omitempty"                                     jsonschema_description:"表单标题"`
	Description string         `json:"description,omitempty"                               jsonschema_description:"表单说明，最多 200 字"`
	Fields      []FormField    `json:"fields"      validate:"required,min=1,dive"          jsonschema:"required" jsonschema_description:"1~20 个字段"`
	SubmitLabel string         `json:"submit_label,omitempty"                              jsonschema_description:"提交按钮文案，默认“提交”"`
	ActionName  string         `json:"action_name" validate:"required"                     jsonschema:"required" jsonschema_description:"提交时回传的事件名，小写下划线，例如 query_minions"`
	Context     map[string]any `json:"context,omitempty"                                   jsonschema_description:"随提交一起回传的上下文，例如 id、slug"`
	SurfaceID   string         `json:"surface_id,omitempty"                                jsonschema_description:"仅整体替换已有表单时传入；一般留空创建新画布"`
}

func (ui *uiInstance) form(ctx context.Context, in RenderFormInput) (*RenderOutput, error) {
	if err := ui.sess.Valid(in); err != nil {
		return nil, err
	}

	surfaceID, err := ui.surfaceID(in.SurfaceID, "form")
	if err != nil {
		return nil, err
	}

	ops, err := buildFormOps(ctx, surfaceID, in)
	if err != nil {
		return nil, err
	}

	return ui.snapshot(ctx, surfaceID, ops)
}

func buildFormOps(_ context.Context, surfaceID string, in RenderFormInput) ([]a2ui.AgentMessage, error) {
	values, err := validateFormInput(&in)
	if err != nil {
		return nil, err
	}

	children := make([]string, 0, len(in.Fields)+4)
	components := make([]a2ui.Component, 0, len(in.Fields)*2+8)

	if title := strings.TrimSpace(in.Title); title != "" {
		children = append(children, "form_title")
		components = append(components, text("form_title", title))
	}
	if desc := strings.TrimSpace(in.Description); desc != "" {
		children = append(children, "form_desc")
		components = append(components, caption("form_desc", desc))
	}

	for i := range in.Fields {
		field := in.Fields[i]
		id := fmt.Sprintf("form_field_%d", i)
		children = append(children, id)

		component, err := formFieldComponent(id, field)
		if err != nil {
			return nil, err
		}
		components = append(components, component)
	}

	submitLabel := strings.TrimSpace(in.SubmitLabel)
	if submitLabel == "" {
		submitLabel = "提交"
	}

	actionCtx := dynamicValues(in.Context)
	if actionCtx == nil {
		actionCtx = make(map[string]a2ui.DynamicValue, len(in.Fields))
	}
	for _, field := range in.Fields {
		actionCtx[field.Name] = a2ui.ValueBinding(pointerJoin("/form", field.Name))
	}

	children = append(children, "form_actions")
	components = append(components,
		row("form_actions", "form_submit"),
		button("form_submit", "form_submit_label", eventAction(in.ActionName, actionCtx), a2ui.ButtonVariantPrimary),
		text("form_submit_label", submitLabel),
	)

	components = append([]a2ui.Component{
		card("root", "form_col"),
		column("form_col", children...),
	}, components...)

	return createSurface(surfaceID, components, map[string]any{"form": values}), nil
}

func formFieldComponent(id string, field FormField) (a2ui.Component, error) {
	path := pointerJoin("/form", field.Name)

	var component a2ui.Component

	switch field.Type {
	case "text", "textarea", "number":
		variant := a2ui.TextFieldVariantShortText
		switch field.Type {
		case "textarea":
			variant = a2ui.TextFieldVariantLongText
		case "number":
			variant = a2ui.TextFieldVariantNumber
		}

		value := a2ui.StringBinding(path)
		component = a2ui.Component{
			ID: id,
			TextField: &a2ui.TextFieldComponent{
				Label:   a2ui.StringLiteral(field.Label),
				Value:   &value,
				Variant: variant,
			},
		}
		if placeholder := strings.TrimSpace(field.Placeholder); placeholder != "" {
			ph := a2ui.StringLiteral(placeholder)
			component.TextField.Placeholder = &ph
		}
	case "select", "multiselect":
		options := make([]a2ui.ChoiceOption, 0, len(field.Options))
		for _, option := range field.Options {
			options = append(options, a2ui.ChoiceOption{
				Label: a2ui.StringLiteral(option.Label),
				Value: option.Value,
			})
		}

		variant := a2ui.ChoicePickerVariantMutuallyExclusive
		if field.Type == "multiselect" {
			variant = a2ui.ChoicePickerVariantMultipleSelection
		}

		label := a2ui.StringLiteral(field.Label)
		component = a2ui.Component{
			ID: id,
			ChoicePicker: &a2ui.ChoicePickerComponent{
				Label:   &label,
				Options: options,
				Value:   a2ui.StringListBinding(path),
				Variant: variant,
			},
		}
	case "checkbox":
		component = a2ui.Component{
			ID: id,
			CheckBox: &a2ui.CheckBoxComponent{
				Label: a2ui.StringLiteral(field.Label),
				Value: a2ui.BoolBinding(path),
			},
		}
	case "slider":
		min := 0.0
		if field.Min != nil {
			min = *field.Min
		}
		max := 100.0
		if field.Max != nil {
			max = *field.Max
		}
		if max <= min {
			return a2ui.Component{}, fmt.Errorf("字段 %q 的 max 必须大于 min", field.Name)
		}

		slider := &a2ui.SliderComponent{
			Label: labelPtr(field.Label),
			Min:   float64Ptr(min),
			Max:   max,
			Value: a2ui.NumberBinding(path),
		}
		if field.Step != nil {
			if *field.Step <= 0 {
				return a2ui.Component{}, fmt.Errorf("字段 %q 的 step 必须大于 0", field.Name)
			}
			steps := int(math.Round((max - min) / *field.Step))
			if steps > 0 {
				slider.Steps = intPtr(steps)
			}
		}
		component = a2ui.Component{ID: id, Slider: slider}
	case "date", "datetime":
		component = a2ui.Component{
			ID: id,
			DateTimeInput: &a2ui.DateTimeInputComponent{
				Label:      labelPtr(field.Label),
				Value:      a2ui.StringBinding(path),
				EnableDate: boolPtr(true),
				EnableTime: boolPtr(field.Type == "datetime"),
			},
		}
	default:
		return a2ui.Component{}, fmt.Errorf("字段 %q 的 type=%q 不支持", field.Name, field.Type)
	}

	if field.Required {
		component.Checks = append(component.Checks, a2ui.CheckRule{
			Condition: a2ui.ValidationFunc(a2ui.FunctionCall{Call: "required"}),
			Message:   "必填",
		})
	}

	return component, nil
}

func validateFormInput(input *RenderFormInput) (map[string]any, error) {
	if input == nil {
		return nil, errors.New("form input is nil")
	}

	if n := len(input.Fields); n == 0 || n > 20 {
		return nil, errors.New("fields 数量必须在 1~20 之间")
	}

	names := make(map[string]struct{}, len(input.Fields))
	values := make(map[string]any, len(input.Fields))

	for i := range input.Fields {
		field := &input.Fields[i]
		field.Name = strings.TrimSpace(field.Name)
		field.Label = strings.TrimSpace(field.Label)

		if !identifierPattern.MatchString(field.Name) {
			return nil, fmt.Errorf("fields[%d].name=%q 不合法，只能包含字母、数字和下划线，且不能以数字开头", i, field.Name)
		}
		if _, exists := names[field.Name]; exists {
			return nil, fmt.Errorf("fields[%d].name=%q 重复", i, field.Name)
		}
		names[field.Name] = struct{}{}

		if field.Label == "" {
			return nil, fmt.Errorf("fields[%d].label 不能为空", i)
		}

		switch field.Type {
		case "select", "multiselect":
			if len(field.Options) == 0 {
				return nil, fmt.Errorf("fields[%d] %q 是 %s，必须提供 options", i, field.Name, field.Type)
			}
			for j, option := range field.Options {
				if strings.TrimSpace(option.Label) == "" || strings.TrimSpace(option.Value) == "" {
					return nil, fmt.Errorf("fields[%d].options[%d] 的 label 和 value 都不能为空", i, j)
				}
			}
		}

		value, err := defaultValue(field)
		if err != nil {
			return nil, fmt.Errorf("fields[%d] %q: %w", i, field.Name, err)
		}
		values[field.Name] = value
	}

	input.ActionName = strings.TrimSpace(input.ActionName)
	if input.ActionName == "" {
		return nil, errors.New("action_name 不能为空")
	}
	if !identifierPattern.MatchString(input.ActionName) {
		return nil, errors.New("action_name 只能包含字母、数字和下划线，且不能以数字开头")
	}
	if err := validateActionContext(input.Context); err != nil {
		return nil, err
	}

	return values, nil
}

func defaultValue(field *FormField) (any, error) {
	switch field.Type {
	case "text", "textarea", "number", "date", "datetime":
		if field.Default == nil {
			return "", nil
		}
		switch v := field.Default.(type) {
		case string:
			return v, nil
		case float64:
			return formatNumber(v), nil
		case bool:
			return strconv.FormatBool(v), nil
		default:
			return nil, errors.New("default 必须是字符串或数字")
		}
	case "select":
		if field.Default == nil {
			return []string{}, nil
		}
		switch v := field.Default.(type) {
		case string:
			if v == "" {
				return []string{}, nil
			}
			return []string{v}, nil
		case float64:
			return []string{formatNumber(v)}, nil
		default:
			return nil, errors.New("default 必须是字符串")
		}
	case "multiselect":
		if field.Default == nil {
			return []string{}, nil
		}
		return stringSlice(field.Default)
	case "checkbox":
		if field.Default == nil {
			return false, nil
		}
		v, ok := field.Default.(bool)
		if !ok {
			return nil, errors.New("default 必须是布尔值")
		}
		return v, nil
	case "slider":
		value := 0.0
		if field.Default != nil {
			v, err := toFloat(field.Default)
			if err != nil {
				return nil, errors.New("default 必须是数字")
			}
			value = v
		}
		min := 0.0
		if field.Min != nil {
			min = *field.Min
		}
		max := 100.0
		if field.Max != nil {
			max = *field.Max
		}
		if max <= min {
			return nil, errors.New("max 必须大于 min")
		}
		if value < min || value > max {
			return nil, fmt.Errorf("default=%v 必须位于 [%v, %v] 之间", value, min, max)
		}
		return value, nil
	default:
		return nil, fmt.Errorf("type=%q 不支持", field.Type)
	}
}

func stringSlice(value any) ([]string, error) {
	switch v := value.(type) {
	case []string:
		return v, nil
	case []any:
		out := make([]string, 0, len(v))
		for i, item := range v {
			text, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("default[%d] 必须是字符串", i)
			}
			out = append(out, text)
		}
		return out, nil
	case string:
		return []string{v}, nil
	default:
		return nil, errors.New("default 必须是字符串数组")
	}
}

func labelPtr(value string) *a2ui.DynamicString {
	label := a2ui.StringLiteral(value)
	return &label
}
