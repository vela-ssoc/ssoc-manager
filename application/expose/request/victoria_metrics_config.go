package request

type VictoriaMetricsConfigCreate struct {
	Name     string `json:"name"     validate:"required,lte=20"`
	Enabled  bool   `json:"enabled"`
	Method   string `json:"method"   validate:"omitempty,oneof=GET POST PUT"`
	URL      string `json:"url"      validate:"required,http_url,lte=200"`
	Username string `json:"username" validate:"lte=100"`
	Password string `json:"password" validate:"lte=100"`
}

type VictoriaMetricsConfigUpdate struct {
	Int64ID
	VictoriaMetricsConfigCreate
}
