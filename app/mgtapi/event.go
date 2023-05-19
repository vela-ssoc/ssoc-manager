package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func Event(svc service.EventService) route.Router {
	table := dynsql.Builder().Build()
	return &eventREST{
		svc:   svc,
		table: table,
	}
}

type eventREST struct {
	svc   service.EventService
	table dynsql.Table
}

func (evt *eventREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/event/cond").GET(evt.Cond)
	bearer.Route("/events").GET(evt.Page)
}

func (evt *eventREST) Cond(c *ship.Context) error {
	res := evt.table.Schema()
	return c.JSON(http.StatusOK, res)
}

func (evt *eventREST) Page(c *ship.Context) error {
	var req param.PageSQL
	if err := c.BindQuery(&req); err != nil {
		return err
	}
	scope, err := evt.table.Inter(req.Input)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}
	page := req.Pager()
	ctx := c.Request().Context()

	count, dats := evt.svc.Page(ctx, page, scope)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}
