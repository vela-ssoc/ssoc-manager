package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/xgfone/ship/v5"
)

func MinionAccount(svc service.MinionAccountService) route.Router {
	return &minionAccountREST{
		svc: svc,
	}
}

type minionAccountREST struct {
	svc   service.MinionAccountService
	table dynsql.Table
}

func (rest *minionAccountREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/accounts").Data(route.Ignore()).GET(rest.Page)
}

func (rest *minionAccountREST) Cond(c *ship.Context) error {
	res := rest.table.Schema()
	return c.JSON(http.StatusOK, res)
}

func (rest *minionAccountREST) Page(c *ship.Context) error {
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
