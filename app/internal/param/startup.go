package param

import "github.com/vela-ssoc/ssoc-common-mb/dal/model"

type StartupUpdate struct {
	ID     int64               `json:"id,string" validate:"required"`
	Logger model.StartupLogger `json:"logger"`
}

type StartupFallbackUpdate struct {
	Logger model.StartupLogger `json:"logger"`
}
