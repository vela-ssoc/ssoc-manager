package request

import "github.com/vela-ssoc/ssoc-common-mb/datalayer/model"

type BrokerUpsert struct {
	Slug    string                `json:"slug"    validate:"required,gte=1,lte=20"`
	Name    string                `json:"name"    validate:"required,gte=2,lte=20"`
	Exposes model.ExposeAddresses `json:"exposes" validate:"required,gte=1,lte=20,dive"`
	Config  model.BrokerConfig    `json:"config"`
}
