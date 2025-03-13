package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/xgfone/ship/v5"
)

func Shared(svc service.SharedService) route.Router {
	filters := []dynsql.Column{
		dynsql.StringColumn("bucket", "bucket").Build(),
		dynsql.StringColumn("key", "key").Build(),
		dynsql.StringColumn("`value`", "value").Build(),
		dynsql.IntColumn("`count`", "count").Build(),
	}
	tbl := dynsql.Builder().Filters(filters...).Build()

	return &sharedREST{
		svc: svc,
		tbl: tbl,
	}
}

type sharedREST struct {
	svc service.SharedService
	tbl dynsql.Table
}

func (rest *sharedREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/shared/strings/cond").
		Data(route.Ignore()).GET(rest.Cond)
	bearer.Route("/shared/strings/sweep").
		Data(route.Named("清理过期数据")).DELETE(rest.Sweep)
	bearer.Route("/shared/strings/keys").
		Data(route.Ignore()).GET(rest.Keys)
	bearer.Route("/shared/strings/buckets").
		Data(route.Ignore()).GET(rest.Buckets)
	bearer.Route("/shared/strings/audits").
		Data(route.Ignore()).GET(rest.Audits)
	bearer.Route("/shared/strings/update").
		Data(route.Ignore()).POST(rest.Update)
}

func (rest *sharedREST) Cond(c *ship.Context) error {
	res := rest.tbl.Schema()
	return c.JSON(http.StatusOK, res)
}

func (rest *sharedREST) Keys(c *ship.Context) error {
	var req param.PageSQL
	if err := c.BindQuery(&req); err != nil {
		return err
	}
	scope, err := rest.tbl.Inter(req.Input)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}
	page := req.Pager()
	ctx := c.Request().Context()

	count, dats := rest.svc.Keys(ctx, page, scope)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}

func (rest *sharedREST) Sweep(c *ship.Context) error {
	req := new(param.SharedBucketKey)
	if err := c.BindQuery(req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Sweep(ctx, req.Bucket, req.Key)
}

func (rest *sharedREST) Buckets(c *ship.Context) error {
	ctx := c.Request().Context()
	ret := rest.svc.Buckets(ctx)

	return c.JSON(http.StatusOK, ret)
}

func (rest *sharedREST) Audits(c *ship.Context) error {
	req := new(param.SharedAuditPage)
	if err := c.BindQuery(req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	p := req.Pager()
	count, dats := rest.svc.Audits(ctx, p, req.Bucket, req.Key)
	ret := p.Result(count, dats)

	return c.JSON(http.StatusOK, ret)
}

func (rest *sharedREST) Update(c *ship.Context) error {
	req := new(param.SharedUpdate)
	if err := c.Bind(req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Update(ctx, req)
}
