package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-common-mb/dynsql"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func NewDomain(svc service.DomainService) *Domain {
	enums := []string{"A", "AAAA", "CNAME", "MX", "NS", "TXT", "SRV", "CAA"}
	typeEnums := dynsql.StringEnum().Sames(enums)

	ispCol := dynsql.StringColumn("isp", "运营商").Build()
	rcdCol := dynsql.StringColumn("record", "域名信息").Build()
	typCol := dynsql.StringColumn("type", "解析类型").Enums(typeEnums).Build()
	addrCol := dynsql.StringColumn("addr", "解析地址").Build()
	comCol := dynsql.StringColumn("`comment`", "备注信息").Build()
	oriCol := dynsql.StringColumn("origin", "来源").Build()
	idCol := dynsql.StringColumn("id", "ID").Build()

	table := dynsql.Builder().
		Filters(ispCol, rcdCol, typCol, addrCol, comCol, oriCol, idCol).
		Build()

	return &Domain{
		svc:   svc,
		table: table,
	}
}

type Domain struct {
	svc   service.DomainService
	table dynsql.Table
}

func (rest *Domain) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/domain/cond").Data(route.Ignore()).GET(rest.Cond)
	bearer.Route("/domains").Data(route.Ignore()).GET(rest.Page)
}

func (rest *Domain) Cond(c *ship.Context) error {
	res := rest.table.Schema()
	return c.JSON(http.StatusOK, res)
}

func (rest *Domain) Page(c *ship.Context) error {
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
