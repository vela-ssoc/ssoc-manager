package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-common-mb/dynsql"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func NewRiskIP(svc *service.RiskIP) *RiskIP {
	filters := []dynsql.Column{
		dynsql.StringColumn("ip", "IP地址").Build(),
		dynsql.StringColumn("kind", "风险类型").Build(),
		dynsql.StringColumn("origin", "数据来源").Build(),
		dynsql.TimeColumn("before_at", "有效期").Build(),
	}
	table := dynsql.Builder().Filters(filters...).Build()
	return &RiskIP{
		svc:   svc,
		table: table,
	}
}

type RiskIP struct {
	svc   *service.RiskIP
	table dynsql.Table
}

func (rest *RiskIP) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/riskip/cond").Data(route.Ignore()).GET(rest.Cond)
	bearer.Route("/riskips").Data(route.Ignore()).POST(rest.Page)
	bearer.Route("/riskip").
		Data(route.Named("新建风险 IP")).POST(rest.Create).
		Data(route.Named("修改风险 IP")).PATCH(rest.Update).
		Data(route.Named("导入风险 IP")).PUT(rest.Import).
		Data(route.Named("删除风险 IP")).DELETE(rest.Delete)
}

func (rest *RiskIP) Cond(c *ship.Context) error {
	res := rest.table.Schema()
	return c.JSON(http.StatusOK, res)
}

func (rest *RiskIP) Page(c *ship.Context) error {
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

func (rest *RiskIP) Create(c *ship.Context) error {
	var req param.RiskIPCreate
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Create(ctx, &req)
}

func (rest *RiskIP) Update(c *ship.Context) error {
	var req param.RiskIPUpdate
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Update(ctx, &req)
}

func (rest *RiskIP) Import(c *ship.Context) error {
	var req param.RiskIPImport
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Import(ctx, &req)
}

func (rest *RiskIP) Delete(c *ship.Context) error {
	var req param.OptionalIDs
	if err := c.Bind(&req); err != nil || len(req.ID) == 0 {
		return err
	}
	ctx := c.Request().Context()

	return rest.svc.Delete(ctx, req.ID)
}
