package param

type TagUpdate struct {
	ID   int64    `json:"id,string" validate:"required"`
	Tags []string `json:"tags"      validate:"gt=0,unique,dive,tag"`
}

type TagSidebar struct {
	IPv4    bool   `query:"ipv4"`                       // 是否显示永久 IPv4 标签
	Keyword string `query:"keyword" validate:"lte=100"` // 搜索参数
}
