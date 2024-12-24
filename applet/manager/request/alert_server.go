package request

type AlertServerUpsert struct {
	Mode    string `json:"mode"    validate:"required,oneof=dong siem"`
	Name    string `json:"name"    validate:"required,lte=20"`
	URL     string `json:"url"     validate:"required,lte=255,http_url"`
	Token   string `json:"token"   validate:"required,lte=255"`
	Account string `json:"account" validate:"lte=20"`
}
