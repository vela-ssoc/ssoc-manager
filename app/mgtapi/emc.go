package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-common-mb/param/request"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/vela-ssoc/ssoc-manager/param/mrequest"
	"github.com/xgfone/ship/v5"
)

func NewEmc(svc service.EmcService) *Emc {
	return &Emc{
		svc: svc,
	}
}

type Emc struct {
	svc service.EmcService
}

func (rest *Emc) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/emcs").Data(route.Ignore()).GET(rest.Page)
	bearer.Route("/emc").
		Data(route.Named("新则咚咚服务号")).POST(rest.Create).
		Data(route.Named("修改咚咚服务号")).PUT(rest.Update).
		Data(route.Named("删除咚咚服务号")).DELETE(rest.Delete)
}

func (rest *Emc) Page(c *ship.Context) error {
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

func (rest *Emc) Create(c *ship.Context) error {
	var req mrequest.EmcCreate
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Create(ctx, &req)
}

func (rest *Emc) Update(c *ship.Context) error {
	var req mrequest.EmcUpdate
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Update(ctx, &req)
}

func (rest *Emc) Delete(c *ship.Context) error {
	var req request.Int64ID
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Delete(ctx, req.ID)
}
