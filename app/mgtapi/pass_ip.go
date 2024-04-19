package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-common-mb-itai/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func PassIP(svc service.PassIPService) route.Router {
	tbl := dynsql.Builder().Filters(
		dynsql.StringColumn("domain", "域名").Build(),
		dynsql.StringColumn("kind", "数据维度").Build(),
		dynsql.TimeColumn("before_at", "有效期").Build(),
	).Build()
	return &passIPREST{
		svc: svc,
		tbl: tbl,
	}
}

type passIPREST struct {
	svc service.PassIPService
	tbl dynsql.Table
}

func (rest *passIPREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/passip/cond").Data(route.Ignore()).GET(rest.Cond)
	bearer.Route("/passips").Data(route.Ignore()).POST(rest.Page)
}

func (rest *passIPREST) Cond(c *ship.Context) error {
	res := rest.tbl.Schema()
	return c.JSON(http.StatusOK, res)
}

func (rest *passIPREST) Page(c *ship.Context) error {
	var req param.PageSQL
	if err := c.Bind(&req); err != nil {
		return err
	}
	scope, err := rest.tbl.Inter(req.Input)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}

	page := req.Pager()
	ctx := c.Request().Context()
	count, dats := rest.svc.Page(ctx, page, scope)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}
