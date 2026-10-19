package restapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-common-mb/profile"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/xgfone/ship/v5"
)

type Config struct {
	cfg *profile.ManagerConfig
}

func NewConfig(cfg *profile.ManagerConfig) *Config {
	return &Config{cfg: cfg}
}

func (cf *Config) RegisterHTTP(g shipx.RouteGroup) error {
	g.APIV3.Route("/config").GET(cf.get) // 中继节点列表
	return nil
}

func (cf *Config) get(c *ship.Context) error {
	return c.JSON(http.StatusOK, cf.cfg)
}
