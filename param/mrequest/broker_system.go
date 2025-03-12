package mrequest

import (
	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/param/request"
)

type BrokerSystemUpdate struct {
	request.Int64ID
	Semver model.Semver `json:"semver" query:"semver" validate:"omitempty,semver"`
}
