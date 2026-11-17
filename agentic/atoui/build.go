package atoui

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/tmc/a2ui"
)

var identifierPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func newSurfaceID(prefix string) (string, error) {
	if prefix == "" {
		prefix = "surface"
	}
	if !identifierPattern.MatchString(prefix) {
		return "", fmt.Errorf("atoui: invalid surface id prefix %q", prefix)
	}

	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}

	return prefix + "_" + hex.EncodeToString(buf[:]), nil
}

func createSurface(
	surfaceID string,
	components []a2ui.Component,
	dataModel map[string]any,
) []a2ui.AgentMessage {
	return []a2ui.AgentMessage{{
		Version: a2ui.Version,
		CreateSurface: &a2ui.CreateSurface{
			SurfaceID:  surfaceID,
			CatalogID:  CatalogID,
			Components: components,
			DataModel:  dataModel,
		},
	}}
}

func text(id, value string) a2ui.Component {
	return a2ui.Component{
		ID: id,
		Text: &a2ui.TextComponent{
			Text: a2ui.StringLiteral(value),
		},
	}
}

func caption(id, value string) a2ui.Component {
	return a2ui.Component{
		ID: id,
		Text: &a2ui.TextComponent{
			Text:    a2ui.StringLiteral(value),
			Variant: a2ui.TextVariantCaption,
		},
	}
}

func textBinding(id, path string) a2ui.Component {
	return a2ui.Component{
		ID: id,
		Text: &a2ui.TextComponent{
			Text: a2ui.StringBinding(path),
		},
	}
}

func captionBinding(id, path string) a2ui.Component {
	return a2ui.Component{
		ID: id,
		Text: &a2ui.TextComponent{
			Text:    a2ui.StringBinding(path),
			Variant: a2ui.TextVariantCaption,
		},
	}
}

func card(id, child string) a2ui.Component {
	return a2ui.Component{
		ID:   id,
		Card: &a2ui.CardComponent{Child: child},
	}
}

func column(id string, children ...string) a2ui.Component {
	return a2ui.Component{
		ID: id,
		Column: &a2ui.ColumnComponent{
			Children: a2ui.ChildList{IDs: children},
		},
	}
}

func row(id string, children ...string) a2ui.Component {
	return a2ui.Component{
		ID: id,
		Row: &a2ui.RowComponent{
			Children: a2ui.ChildList{IDs: children},
		},
	}
}

func listTemplate(id, path, componentID string) a2ui.Component {
	return a2ui.Component{
		ID: id,
		List: &a2ui.ListComponent{
			Children: a2ui.ChildList{
				Template: &a2ui.ChildTemplate{
					ComponentID: componentID,
					Path:        path,
				},
			},
		},
	}
}

func divider(id string) a2ui.Component {
	return a2ui.Component{
		ID:      id,
		Divider: &a2ui.DividerComponent{},
	}
}

func icon(id string, name a2ui.IconName) a2ui.Component {
	return a2ui.Component{
		ID: id,
		Icon: &a2ui.IconComponent{
			Name: a2ui.IconNameOrPath{Name: &name},
		},
	}
}

func iconBinding(id, path string) a2ui.Component {
	return a2ui.Component{
		ID: id,
		Icon: &a2ui.IconComponent{
			Name: a2ui.IconNameOrPath{
				Binding: &a2ui.DataBinding{Path: path},
			},
		},
	}
}

func button(id, child string, action a2ui.Action, variant a2ui.ButtonVariant) a2ui.Component {
	return a2ui.Component{
		ID: id,
		Button: &a2ui.ButtonComponent{
			Child:   child,
			Action:  action,
			Variant: variant,
		},
	}
}

func eventAction(name string, ctx map[string]a2ui.DynamicValue) a2ui.Action {
	return a2ui.Action{
		Event: &a2ui.EventAction{
			Name:    name,
			Context: ctx,
		},
	}
}

func weight(c a2ui.Component, value float64) a2ui.Component {
	c.Weight = &value
	return c
}

func dynamicValues(values map[string]any) map[string]a2ui.DynamicValue {
	if len(values) == 0 {
		return nil
	}

	out := make(map[string]a2ui.DynamicValue, len(values))
	for key, value := range values {
		switch val := value.(type) {
		case string:
			out[key] = a2ui.ValueString(val)
		case float64:
			out[key] = a2ui.ValueNumber(val)
		case float32:
			out[key] = a2ui.ValueNumber(float64(val))
		case int:
			out[key] = a2ui.ValueNumber(float64(val))
		case int8:
			out[key] = a2ui.ValueNumber(float64(val))
		case int16:
			out[key] = a2ui.ValueNumber(float64(val))
		case int32:
			out[key] = a2ui.ValueNumber(float64(val))
		case int64:
			out[key] = a2ui.ValueNumber(float64(val))
		case uint:
			out[key] = a2ui.ValueNumber(float64(val))
		case uint8:
			out[key] = a2ui.ValueNumber(float64(val))
		case uint16:
			out[key] = a2ui.ValueNumber(float64(val))
		case uint32:
			out[key] = a2ui.ValueNumber(float64(val))
		case uint64:
			out[key] = a2ui.ValueNumber(float64(val))
		case bool:
			out[key] = a2ui.ValueBool(val)
		case []any:
			out[key] = a2ui.ValueArray(val)
		case []string:
			items := make([]any, 0, len(val))
			for _, item := range val {
				items = append(items, item)
			}
			out[key] = a2ui.ValueArray(items)
		}
	}

	return out
}

func validateActionContext(values map[string]any) error {
	for key, value := range values {
		if strings.TrimSpace(key) == "" {
			return errors.New("context 的 key 不能为空")
		}

		switch val := value.(type) {
		case string, float64, float32, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, bool:
		case []any:
			for i, item := range val {
				switch item.(type) {
				case string, float64, float32, int, int64, bool:
				default:
					return fmt.Errorf("context[%q][%d] 只支持字符串、数字或布尔", key, i)
				}
			}
		default:
			return fmt.Errorf("context[%q] 只支持字符串、数字、布尔或由它们组成的数组", key)
		}
	}

	return nil
}

func pointerEscape(key string) string {
	key = strings.ReplaceAll(key, "~", "~0")
	return strings.ReplaceAll(key, "/", "~1")
}

func pointerJoin(path, key string) string {
	if path == "" || path == "/" {
		return "/" + pointerEscape(key)
	}

	return strings.TrimSuffix(path, "/") + "/" + pointerEscape(key)
}

func float64Ptr(value float64) *float64 {
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}

func intPtr(value int) *int {
	return &value
}
