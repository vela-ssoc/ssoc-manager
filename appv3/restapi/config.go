package restapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/vela-ssoc/ssoc-manager/config"
	"github.com/xgfone/ship/v5"
)

type Config struct {
	boot config.Config
	cfg  *model.ManagerConfig
}

func NewConfig(boot config.Config, cfg *model.ManagerConfig) *Config {
	return &Config{boot: boot, cfg: cfg}
}

func (cf *Config) RegisterHTTP(g shipx.RouteGroup3) error {
	g.ExposeV3.Group.Route("/config").GET(cf.get) // 中继节点列表
	return nil
}

func (cf *Config) get(c *ship.Context) error {
	return c.JSON(http.StatusOK, cf.cfg)
}
