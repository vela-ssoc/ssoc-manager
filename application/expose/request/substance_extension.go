package request

import "encoding/json"

type SubstanceExtensionCreate struct {
	Name        string          `json:"name"                validate:"substance_name"` // 配置名字
	MinionID    int64           `json:"minion_id,string"`                              // 节点 ID
	ExtensionID int64           `json:"extension_id,string" validate:"required"`       // 插件 ID
	Data        json.RawMessage `json:"data"`                                          // 渲染参数
}

type SubstanceExtensionUpdate struct {
	ID   int64           `json:"id,string"           validate:"required"` // 配置 ID
	Data json.RawMessage `json:"data"`                                    // 渲染参数
	// ExtensionID int64           `json:"extension_id,string" validate:"required"` // 插件 ID
}
