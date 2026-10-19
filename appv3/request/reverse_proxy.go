package request

import "github.com/vela-ssoc/ssoc-common-mb/datalayer/model"

type ReverseProxyUpsert struct {
	Location string           `json:"location" validate:"required"`
	Name     string           `json:"name"     validate:"required,lte=20"`
	Upstream string           `json:"upstream" validate:"required,http_url"`
	Header   model.HTTPHeader `json:"header"   validate:"lte=20"`
	Manager  bool             `json:"manager"`
	Broker   bool             `json:"broker"`
}

type ReverseProxyLocation struct {
	Location string `json:"location" query:"location" validate:"required"`
}
