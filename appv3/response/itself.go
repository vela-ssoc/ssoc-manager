package response

import (
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
	"github.com/vela-ssoc/ssoc-manager/config"
)

type ItselfConfig struct {
	Boot   config.Config        `json:"boot"`
	Config *model.ManagerConfig `json:"config"`
}
