package request

import (
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
)

type ItselfExit struct {
	Seconds int `json:"seconds" query:"seconds" validate:"gte=0"`
}

func (ie ItselfExit) Duration() time.Duration {
	return time.Duration(ie.Seconds) * time.Second
}

type ItselfConfigUpdate struct {
	Server *model.ManagerServerConfig `json:"server"`
	Logger *model.LoggerConfig        `json:"logger"`
}
