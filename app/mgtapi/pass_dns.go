package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-common-mb/dynsql"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func NewPassDNS(svc service.PassDNSService) *PassDNS {
	tbl := dynsql.Builder().Filters(
		dynsql.StringColumn("domain", "域名").Build(),
		dynsql.StringColumn("kind", "数据维度").Build(),
		dynsql.TimeColumn("before_at", "有效期").Build(),
	).Build()
	return &PassDNS{
		svc: svc,
		tbl: tbl,
	}
}

type PassDNS struct {
	svc service.PassDNSService
	tbl dynsql.Table
}

func (rest *PassDNS) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/passdns/cond").Data(route.Ignore()).GET(rest.Cond)
	bearer.Route("/passdnss").Data(route.Ignore()).POST(rest.Page)
}

func (rest *PassDNS) Cond(c *ship.Context) error {
	res := rest.tbl.Schema()
	return c.JSON(http.StatusOK, res)
}

func (rest *PassDNS) Page(c *ship.Context) error {
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
