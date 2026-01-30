package request

type AgentConsoleRead struct {
	ID   int64  `json:"id"   form:"id"   query:"id"   validate:"required"` // agent 节点 ID
	From string `json:"from" form:"from" query:"from" validate:"required"` // 模块
	N    int    `json:"n"    form:"n"    query:"n"`
}

type AgentConsoleRemove struct {
	ID   int64  `json:"id"   form:"id"   query:"id"   validate:"required"` // agent 节点 ID
	From string `json:"from" form:"from" query:"from" validate:"required"` // 模块
}

type AgentConsoleStat struct {
	Size    int64 `json:"size"`
	Maxsize int64 `json:"maxsize"`
}
