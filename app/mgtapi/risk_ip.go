package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func RiskIP(svc service.RiskIPService) route.Router {
	filters := []dynsql.Column{
		dynsql.StringColumn("ip", "IP地址").Build(),
		dynsql.StringColumn("kind", "风险类型").Build(),
		dynsql.StringColumn("origin", "数据来源").Build(),
		dynsql.TimeColumn("before_at", "有效期").Build(),
	}
	table := dynsql.Builder().Filters(filters...).Build()
	return &riskIPREST{
		svc:   svc,
		table: table,
	}
}

type riskIPREST struct {
	svc   service.RiskIPService
	table dynsql.Table
}

func (rest *riskIPREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/riskip/cond").GET(rest.Cond)
	bearer.Route("/riskips").POST(rest.Page)
	bearer.Route("/riskip").
		POST(rest.Create).
		PATCH(rest.Update).
		PUT(rest.Import).
		DELETE(rest.Delete)
}

func (rest *riskIPREST) Cond(c *ship.Context) error {
	res := rest.table.Schema()
	return c.JSON(http.StatusOK, res)
}

func (rest *riskIPREST) Page(c *ship.Context) error {
	var req param.PageSQL
	if err := c.Bind(&req); err != nil {
		return err
	}
	scope, err := rest.table.Inter(req.Input)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}
	page := req.Pager()
	ctx := c.Request().Context()

	count, dats := rest.svc.Page(ctx, page, scope)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}

func (rest *riskIPREST) Create(c *ship.Context) error {
	var req param.RiskIPCreate
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Create(ctx, &req)
}

func (rest *riskIPREST) Update(c *ship.Context) error {
	var req param.RiskIPUpdate
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Update(ctx, &req)
}

func (rest *riskIPREST) Import(c *ship.Context) error {
	var req param.RiskIPImport
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Import(ctx, &req)
}

func (rest *riskIPREST) Delete(c *ship.Context) error {
	var req param.OptionalIDs
	if err := c.Bind(&req); err != nil || len(req.ID) == 0 {
		return err
	}
	ctx := c.Request().Context()

	return rest.svc.Delete(ctx, req.ID)
}
