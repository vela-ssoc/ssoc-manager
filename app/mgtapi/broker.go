package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func Broker(svc service.BrokerService) route.Router {
	return &brokerREST{
		svc: svc,
	}
}

type brokerREST struct {
	svc service.BrokerService
}

func (rest *brokerREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/brokers").Data(route.Ignore()).GET(rest.Page)
	bearer.Route("/broker/indices").Data(route.Ignore()).GET(rest.Indices)
}

func (rest *brokerREST) Page(c *ship.Context) error {
	var req param.Page
	if err := c.BindQuery(&req); err != nil {
		return err
	}
	page := req.Pager()
	ctx := c.Request().Context()

	count, dats := rest.svc.Page(ctx, page)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}

func (rest *brokerREST) Indices(c *ship.Context) error {
	var req param.Index
	if err := c.BindQuery(&req); err != nil {
		return err
	}
	idx := req.Indexer()
	ctx := c.Request().Context()
	dats := rest.svc.Indices(ctx, idx)

	return c.JSON(http.StatusOK, dats)
}
