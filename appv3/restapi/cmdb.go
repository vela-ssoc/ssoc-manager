package restapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"github.com/vela-ssoc/ssoc-manager/appv3/response"
	"github.com/vela-ssoc/ssoc-manager/appv3/service"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/xgfone/ship/v5"
)

type CMDB struct {
	svc *service.CMDB
}

func NewCMDB(svc *service.CMDB) *CMDB {
	return &CMDB{svc: svc}
}

func (cdb *CMDB) RegisterHTTP(g shipx.RouteGroup3) error {
	g.ExposeV3.Group.Route("/cmdb/departments").GET(cdb.departments)
	g.ExposeV3.Group.Route("/cmdb/assets").GET(cdb.assets)

	return nil
}

func (cdb *CMDB) departments(c *ship.Context) error {
	ctx := c.Request().Context()
	nodes, err := cdb.svc.Department(ctx)
	if err != nil {
		return err
	}
	ret := response.NewRecords(nodes)

	return c.JSON(http.StatusOK, ret)
}

func (cdb *CMDB) assets(c *ship.Context) error {
	var param request.CMDBAssets
	if err := c.BindQuery(&param); err != nil {
		return err
	}

	ctx := c.Request().Context()
	ret, err := cdb.svc.Assets(ctx, param)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}
