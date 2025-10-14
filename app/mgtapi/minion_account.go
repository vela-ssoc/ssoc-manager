package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-common-mb/dynsql"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func NewMinionAccount(svc service.MinionAccountService) *MinionAccount {
	return &MinionAccount{
		svc: svc,
	}
}

type MinionAccount struct {
	svc   service.MinionAccountService
	table dynsql.Table
}

func (rest *MinionAccount) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/accounts").Data(route.Ignore()).GET(rest.Page)
}

func (rest *MinionAccount) Cond(c *ship.Context) error {
	res := rest.table.Schema()
	return c.JSON(http.StatusOK, res)
}

func (rest *MinionAccount) Page(c *ship.Context) error {
	var req param.MinionAccountPage
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	page := req.Pager()
	ctx := c.Request().Context()
	count, dats := rest.svc.Page(ctx, page, req.MinionID, req.Name)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}
