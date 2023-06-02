package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func SBOMVuln(svc service.SBOMVulnService) route.Router {
	purlCol := dynsql.StringColumn("purl", "PURL").Build()
	vidCol := dynsql.StringColumn("vuln_id", "漏洞编号").Build()
	table := dynsql.Builder().Filters(purlCol, vidCol).Build()
	return &sbomVulnREST{
		svc:   svc,
		table: table,
	}
}

type sbomVulnREST struct {
	svc   service.SBOMVulnService
	table dynsql.Table
}

func (rest *sbomVulnREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/sbom/vulns").GET(rest.Page)
	bearer.Route("/sbom/vuln/cond").GET(rest.Cond)
	bearer.Route("/sbom/vuln/projects").GET(rest.Project)
}

func (rest *sbomVulnREST) Cond(c *ship.Context) error {
	res := rest.table.Schema()
	return c.JSON(http.StatusOK, res)
}

func (rest *sbomVulnREST) Page(c *ship.Context) error {
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

func (rest *sbomVulnREST) Project(c *ship.Context) error {
	var req param.SBOMVulnProject
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	page := req.Pager()
	ctx := c.Request().Context()
	count, dats := rest.svc.Project(ctx, page, req.PURL)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}
