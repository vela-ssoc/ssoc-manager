package service

import (
	"encoding/json"
	"io"

	"github.com/vela-ssoc/luatemplate"
)

func NewLuaTemplate() *LuaTemplate {
	return &LuaTemplate{}
}

type LuaTemplate struct{}

func (t *LuaTemplate) Preparse(source string) (json.RawMessage, error) {
	tpl, err := luatemplate.New("lua").Parse(source)
	if err != nil {
		return nil, err
	}
	param := tpl.ParamJSON()

	return param, nil
}

func (t *LuaTemplate) Prerender(w io.Writer, source string, data any) error {
	tpl, err := luatemplate.New("lua").Parse(source)
	if err != nil {
		return err
	}

	return tpl.Execute(w, data)
}
