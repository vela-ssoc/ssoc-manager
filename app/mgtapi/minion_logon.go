package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func MinionLogon(svc service.MinionLogonService) route.Router {
	table := dynsql.Builder().Build()
	return &minionLogonREST{
		svc:   svc,
		table: table,
	}
}

type minionLogonREST struct {
	svc   service.MinionLogonService
	table dynsql.Table
}

func (rest *minionLogonREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/logons").GET(rest.Page)
}

func (rest *minionLogonREST) Cond(c *ship.Context) error {
	res := rest.table.Schema()
	return c.JSON(http.StatusOK, res)
}

func (rest *minionLogonREST) Page(c *ship.Context) error {
	var req param.PageSQL
	if err := c.BindQuery(&req); err != nil {
		return err
	}
	scope, err := rest.table.Inter(req.Input)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}

	page := req.Pager()
	ctx := c.Request().Context()
	count, dats := rest.svc.Page(ctx, page, scope)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}
