package param

type AlertServerUpsert struct {
	Mode    string `json:"mode"    binding:"oneof=siem dong"`
	Name    string `json:"name"    binding:"required,lte=20"`
	URL     string `json:"url"     binding:"required,lte=255"`
	Token   string `json:"token"   binding:"required,lte=255"`
	Account string `json:"account" binding:"lte=20"`
}
