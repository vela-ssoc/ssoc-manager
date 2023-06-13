package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/vela-ssoc/vela-manager/app/session"
	"github.com/xgfone/ship/v5"
)

func Compound(svc service.CompoundService) route.Router {
	return &compoundREST{
		svc: svc,
	}
}

type compoundREST struct {
	svc service.CompoundService
}

func (rest *compoundREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/compounds").Data(route.Ignore()).GET(rest.Page)
	bearer.Route("/compound/indices").Data(route.Ignore()).GET(rest.Indices)
	bearer.Route("/compound").
		Data(route.Named("新增配置组合")).POST(rest.Create).
		Data(route.Named("修改配置组合")).PUT(rest.Update).
		Data(route.Named("删除配置组合")).DELETE(rest.Delete)
}

func (rest *compoundREST) Indices(c *ship.Context) error {
	var req param.Index
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	idx := req.Indexer()
	ctx := c.Request().Context()
	res := rest.svc.Indices(ctx, idx)

	return c.JSON(http.StatusOK, res)
}

func (rest *compoundREST) Page(c *ship.Context) error {
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

func (rest *compoundREST) Create(c *ship.Context) error {
	var req param.CompoundCreate
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	cu := session.Cast(c.Any)

	return rest.svc.Create(ctx, &req, cu.ID)
}

func (rest *compoundREST) Update(c *ship.Context) error {
	var req param.CompoundUpdate
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	cu := session.Cast(c.Any)

	return rest.svc.Update(ctx, &req, cu.ID)
}

func (rest *compoundREST) Delete(c *ship.Context) error {
	var req param.IntID
	if err := c.BindQuery(&req); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return rest.svc.Delete(ctx, req.ID)
}
