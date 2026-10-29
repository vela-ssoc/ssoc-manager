package request

import "github.com/vela-ssoc/ssoc-common-mb/datalayer/model"

type VictoriaMetricsConfigUpsert struct {
	Enabled  bool             `json:"enabled"`
	Address  string           `json:"address"   validate:"required,http_url"`
	TenantID string           `json:"tenant_id" validate:"lte=20"`
	Username string           `json:"username"  validate:"lte=100"`
	Password string           `json:"password"  validate:"lte=100"`
	Headers  model.HTTPHeader `json:"headers"   validate:"lte=100"`
}
