package route

import (
	"encoding/json"

	"github.com/xgfone/ship/v5"
)

// Describer 路由描述信息，用于记录请求日志。
type Describer interface {
	// Ignore 是否忽略日志记录
	// 只是简单的信息查看接口，一般都不会记录日志
	Ignore() bool

	// Name 路由接口名字，简短描述接口的功能
	Name(*ship.Context) string

	// Desensitization 对请求 Body 脱敏，如果无需脱敏就原样返回。
	Desensitization([]byte) []byte
}

func Ignore() Describer {
	return &descRoute{ignore: true}
}

func Named(name string) Describer {
	return &descRoute{name: name}
}

func IgnoreBody(name string) Describer {
	dr := &descRoute{name: name}
	dr.dest = dr.empty
	return dr
}

func DestPasswd(name string) Describer {
	dr := &descRoute{name: name}
	dr.dest = dr.password
	return dr
}

type descRoute struct {
	ignore bool
	name   string
	dest   func([]byte) []byte
}

func (dr *descRoute) Ignore() bool              { return dr.ignore }
func (dr *descRoute) Name(*ship.Context) string { return dr.name }

func (dr *descRoute) Desensitization(raw []byte) []byte {
	if dest := dr.dest; dest != nil {
		return dest(raw)
	}
	return raw
}

func (dr *descRoute) empty(raw []byte) []byte {
	return nil
}

func (dr *descRoute) password(raw []byte) []byte {
	keys := []string{"original", "password"}
	mask := "******"
	return dr.desensitizationJSON(keys, mask, raw)
}

// desensitizationJSON 对 JSON 某个字段的内容
// key 为 JSON 的 key，mask 为遮罩层，raw JSON 数据
func (dr *descRoute) desensitizationJSON(keys []string, mask string, raw []byte) []byte {
	var temp map[string]any
	if err := json.Unmarshal(raw, &temp); err != nil {
		return raw
	}
	for _, key := range keys {
		dr.replace(key, mask, temp)
	}

	ret, err := json.Marshal(temp)
	if err != nil {
		return raw
	}

	return ret
}

func (dr *descRoute) replace(key, mask string, mp map[string]any) {
	v, ok := mp[key]
	if !ok || v == nil {
		return
	}
	if s, yes := v.(string); !yes || s != "" {
		mp[key] = mask
	}
}
