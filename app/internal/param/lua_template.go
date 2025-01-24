package param

import "encoding/json"

type LuaTemplatePreparse struct {
	Content string `json:"content" validate:"required,lte=65535"`
}

type LuaTemplatePrerender struct {
	Content string          `json:"content" validate:"required,lte=65535"`
	Data    json.RawMessage `json:"data"    validate:"required,lte=65535"`
}
