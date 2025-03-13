package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/xgfone/ship/v5"
)

func VIP(svc service.VIPService) route.Router {
	vipCol := dynsql.StringColumn("virtual_ip", "公网地址").Build()
	vportCol := dynsql.IntColumn("virtual_port", "公网端口").Build()
	deptCol := dynsql.StringColumn("biz_dept", "业务部门").Build()
	idcCol := dynsql.StringColumn("idc", "IDC").Build()
	batCol := dynsql.TimeColumn("before_at", "有效期").Build()
	table := dynsql.Builder().
		Filters(vipCol, vportCol, deptCol, idcCol, batCol).
		Build()
	return &vipREST{
		svc:   svc,
		table: table,
	}
}

type vipREST struct {
	svc   service.VIPService
	table dynsql.Table
}

func (rest *vipREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/vip/cond").Data(route.Ignore()).GET(rest.Cond)
	bearer.Route("/vips").Data(route.Ignore()).GET(rest.Page)
}

func (rest *vipREST) Cond(c *ship.Context) error {
	res := rest.table.Schema()
	return c.JSON(http.StatusOK, res)
}

func (rest *vipREST) Page(c *ship.Context) error {
	var req param.PageSQL
	if err := c.BindQuery(&req); err != nil {
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
