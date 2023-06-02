package param

type EmcCreate struct {
	Name    string `json:"name"    validate:"required,lte=50"`       // 名字
	Host    string `json:"host"    validate:"required,http,lte=200"` // 咚咚服务器
	Account string `json:"account" validate:"required,lte=200"`      // 咚咚 Account
	Token   string `json:"token"   validate:"required,lte=200"`      // 咚咚 Token
	Enable  bool   `json:"enable"`                                   // 是否启用
}

type EmcUpdate struct {
	ID int64 `json:"id,string" validate:"required"`
	EmcCreate
}
