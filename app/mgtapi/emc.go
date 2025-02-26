package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-common-mb/param/request"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/vela-ssoc/vela-manager/param/mrequest"
	"github.com/xgfone/ship/v5"
)

func Emc(svc service.EmcService) route.Router {
	return &emcREST{
		svc: svc,
	}
}

type emcREST struct {
	svc service.EmcService
}

func (rest *emcREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/emcs").Data(route.Ignore()).GET(rest.Page)
	bearer.Route("/emc").
		Data(route.Named("新则咚咚服务号")).POST(rest.Create).
		Data(route.Named("修改咚咚服务号")).PUT(rest.Update).
		Data(route.Named("删除咚咚服务号")).DELETE(rest.Delete)
}

func (rest *emcREST) Page(c *ship.Context) error {
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

func (rest *emcREST) Create(c *ship.Context) error {
	var req mrequest.EmcCreate
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Create(ctx, &req)
}

func (rest *emcREST) Update(c *ship.Context) error {
	var req mrequest.EmcUpdate
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Update(ctx, &req)
}

func (rest *emcREST) Delete(c *ship.Context) error {
	var req request.Int64ID
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Delete(ctx, req.ID)
}
