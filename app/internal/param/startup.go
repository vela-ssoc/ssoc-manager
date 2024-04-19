package param

import "github.com/vela-ssoc/vela-common-mb-itai/dal/model"

type StartupDetail struct {
	Param   *model.Startup `json:"param"`
	Version int            `json:"version"`
	Chunk   string         `json:"chunk"`
}
