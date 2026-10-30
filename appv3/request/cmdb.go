package request

type CMDBConfigUpsert struct {
	Enabled  bool   `json:"enabled"`
	Endpoint string `json:"endpoint" validate:"required,http_url"`
	Key      string `json:"key"      validate:"required,lte=100"`
	Secret   string `json:"secret"   validate:"required,lte=100"`
}

type CMDBAssets struct {
	PageSizeKeyword
	CIType     []string `json:"ci_type"    query:"ci_type"    validate:"lte=10,unique,dive,required"`
	Department string   `json:"department" query:"department" validate:"lte=255"`
}
