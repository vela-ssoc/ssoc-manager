package request

import "github.com/vela-ssoc/ssoc-common/store/model"

type BrokerCreate struct {
	Name    string                `json:"name"    validate:"required,lte=20"`
	Exposes model.ExposeAddresses `json:"exposes" validate:"required,lte=10,dive"`
	Config  model.BrokerConfig    `json:"config"`
}
