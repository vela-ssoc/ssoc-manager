package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func Store(svc service.StoreService) route.Router {
	return &storeREST{
		svc: svc,
	}
}

type storeREST struct {
	svc service.StoreService
}

func (rest *storeREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/stores").GET(rest.Page)
}

func (rest *storeREST) Page(c *ship.Context) error {
	var req param.PageSQL
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	page := req.Pager()
	ctx := c.Request().Context()
	count, dats := rest.svc.Page(ctx, page)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}
