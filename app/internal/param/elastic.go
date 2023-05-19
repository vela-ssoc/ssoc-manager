package param

type ElasticCreate struct {
	Host     string `json:"host"     validate:"http"`    // ES 地址
	Username string `json:"username"`                    // ES 用户名
	Password string `json:"password"`                    // ES 密码
	Desc     string `json:"desc"     validate:"lte=100"` // 简介
	Enable   bool   `json:"enable"`                      // 是否选中
}

type ElasticUpdate struct {
	IntID
	ElasticCreate
}
