package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func SBOMProject(svc service.SBOMProjectService) route.Router {
	inetCol := dynsql.StringColumn("inet", "终端IP").Build()
	fpCol := dynsql.StringColumn("filepath", "文件").Build()
	pidCol := dynsql.IntColumn("pid", "PID").Build()
	cnumCol := dynsql.IntColumn("component_num", "组件数").Build()
	exeCol := dynsql.StringColumn("exe", "进程名").Build()
	idCol := dynsql.IntColumn("id", "文件ID").Build()
	midCol := dynsql.IntColumn("minion_id", "终端ID").Build()
	table := dynsql.Builder().
		Filters(inetCol, fpCol, pidCol, cnumCol, exeCol, pidCol, midCol, idCol).
		Build()

	return &sbomProjectREST{
		svc:   svc,
		table: table,
	}
}

type sbomProjectREST struct {
	svc   service.SBOMProjectService
	table dynsql.Table
}

func (rest *sbomProjectREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/sbom/projects").Data(route.Ignore()).GET(rest.Page)
	bearer.Route("/sbom/project/cond").Data(route.Ignore()).GET(rest.Cond)
}

func (rest *sbomProjectREST) Cond(c *ship.Context) error {
	res := rest.table.Schema()
	return c.JSON(http.StatusOK, res)
}

func (rest *sbomProjectREST) Page(c *ship.Context) error {
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
