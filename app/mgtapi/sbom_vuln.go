package mgtapi

import (
	"net/http"
	"strconv"

	"github.com/vela-ssoc/ssoc-common-mb/dynsql"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
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
	bearer.Route("/sbom/vulns").Data(route.Ignore()).GET(rest.Page)
	bearer.Route("/sbom/vuln/cond").Data(route.Ignore()).GET(rest.Cond)
	bearer.Route("/sbom/vuln/projects").Data(route.Ignore()).GET(rest.Project)
	bearer.Route("/vulnerabilities").Data(route.Named("拉漏洞库")).GET(rest.Vulnerabilities)
	bearer.Route("/sbom/purl").Data(route.Named("上报 PURL")).POST(rest.Purl)
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

func (rest *sbomVulnREST) Vulnerabilities(c *ship.Context) error {
	strSize := c.Query("size")
	strOffsetID := c.Query("offset_id")
	size, _ := strconv.Atoi(strSize)
	if size <= 0 || size > 200 {
		size = 200
	}
	id, _ := strconv.ParseInt(strOffsetID, 10, 64)
	ctx := c.Request().Context()

	ret := rest.svc.Vulnerabilities(ctx, id, size)

	return c.JSON(http.StatusOK, ret)
}

func (rest *sbomVulnREST) Purl(c *ship.Context) error {
	req := new(param.ReportPurl)
	if err := c.Bind(req); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return rest.svc.Purl(ctx, req)
}
