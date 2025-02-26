package mrequest

type AlertServerUpsert struct {
	Mode    string `json:"mode"    validate:"oneof=siem dong"`
	Name    string `json:"name"    validate:"required,lte=20"`
	URL     string `json:"url"     validate:"required,lte=255"`
	Token   string `json:"token"   validate:"required,lte=255"`
	Account string `json:"account" validate:"lte=20"`
}
