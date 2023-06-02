package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func MinionBinary(svc service.MinionBinaryService) route.Router {
	return &minionBinaryREST{
		svc: svc,
	}
}

type minionBinaryREST struct {
	svc service.MinionBinaryService
}

func (rest *minionBinaryREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/monbins").GET(rest.Page)
	bearer.Route("/monbin/deprecate").PATCH(rest.Deprecate)
	bearer.Route("/monbin").DELETE(rest.Delete)
}

func (rest *minionBinaryREST) Page(c *ship.Context) error {
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

func (rest *minionBinaryREST) Deprecate(c *ship.Context) error {
	var req param.IntID
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Deprecate(ctx, req.ID)
}

func (rest *minionBinaryREST) Delete(c *ship.Context) error {
	var req param.IntID
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Delete(ctx, req.ID)
}
