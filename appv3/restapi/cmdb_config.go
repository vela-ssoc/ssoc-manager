package restapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"github.com/vela-ssoc/ssoc-manager/appv3/service"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/xgfone/ship/v5"
)

type CMDBConfig struct {
	svc *service.CMDBConfig
}

func NewCMDBConfig(svc *service.CMDBConfig) *CMDBConfig {
	return &CMDBConfig{svc: svc}
}

func (cc *CMDBConfig) RegisterHTTP(g shipx.RouteGroup3) error {
	g.ExposeV3.Group.Route("/cmdb/config").
		GET(cc.get).
		POST(cc.upsert)

	return nil
}

func (cc *CMDBConfig) get(c *ship.Context) error {
	ctx := c.Request().Context()
	ret, _ := cc.svc.Get(ctx)

	return c.JSON(http.StatusOK, ret)
}

func (cc *CMDBConfig) upsert(c *ship.Context) error {
	var param request.CMDBConfigUpsert
	if err := c.Bind(&param); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return cc.svc.Upsert(ctx, param)
}
