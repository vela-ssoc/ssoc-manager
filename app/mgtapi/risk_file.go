package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-common-mb/dynsql"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func NewRiskFile(svc *service.RiskFile) *RiskFile {
	algorithms := []string{"MD2", "MD4", "MD5", "MD6", "SHA1", "SHA224", "SHA256", "SHA384", "SHA512"}
	algorithmEnums := dynsql.StringEnum().Sames(algorithms)

	tbl := dynsql.Builder().Filters(
		dynsql.StringColumn("algorithm", "算法").Enums(algorithmEnums).Build(),
		dynsql.StringColumn("checksum", "哈希").Build(),
		dynsql.StringColumn("kind", "风险类型").Build(),
		dynsql.StringColumn("origin", "数据来源").Build(),
		dynsql.TimeColumn("before_at", "有效期").Build(),
		dynsql.StringColumn("desc", "说明").Build(),
	).Build()

	return &RiskFile{
		svc: svc,
		tbl: tbl,
	}
}

type RiskFile struct {
	svc *service.RiskFile
	tbl dynsql.Table
}

func (rest *RiskFile) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/riskfile/cond").Data(route.Ignore()).GET(rest.Cond)
	bearer.Route("/riskfiles").Data(route.Ignore()).POST(rest.Page)
}

func (rest *RiskFile) Cond(c *ship.Context) error {
	res := rest.tbl.Schema()
	return c.JSON(http.StatusOK, res)
}

func (rest *RiskFile) Page(c *ship.Context) error {
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
