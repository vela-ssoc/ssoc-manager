package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-common-mb-itai/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func RiskDNS(svc service.RiskDNSService) route.Router {
	tbl := dynsql.Builder().
		Filters(
			dynsql.StringColumn("domain", "域名").Build(),
			dynsql.StringColumn("kind", "数据维度").Build(),
			dynsql.TimeColumn("before_at", "有效期").Build(),
		).
		Build()

	return &riskDNSREST{
		svc: svc,
		tbl: tbl,
	}
}

type riskDNSREST struct {
	svc service.RiskDNSService
	tbl dynsql.Table
}

func (rest *riskDNSREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/riskdns/cond").Data(route.Ignore()).GET(rest.Cond)
	bearer.Route("/riskdnss").Data(route.Ignore()).POST(rest.Page)
}

func (rest *riskDNSREST) Cond(c *ship.Context) error {
	res := rest.tbl.Schema()
	return c.JSON(http.StatusOK, res)
}

func (rest *riskDNSREST) Page(c *ship.Context) error {
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
