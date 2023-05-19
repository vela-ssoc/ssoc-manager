package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func SBOM(svc service.SBOMService) route.Router {
	// pts := opencond.Patterns{
	//		opencond.Builder("inet", "inet", "终端IP", opencond.TypeString).Build(),
	//		opencond.Builder("name", "name", "组件名", opencond.TypeString).Build(),
	//		opencond.Builder("version", "version", "版本", opencond.TypeString).Build(),
	//		opencond.Builder("purl", "purl", "PURL", opencond.TypeString).Build(),
	//		opencond.Builder("total_num", "total_num", "漏洞总数", opencond.TypeInt).Build(),
	//		opencond.Builder("project_id", "project_id", "文件ID", opencond.TypeInt).Build(),
	//		opencond.Builder("minion_id", "minion_id", "终端ID", opencond.TypeInt).Build(),
	//		opencond.Builder("id", "id", "组件ID", opencond.TypeInt).Build(),
	//	}
	idCol := dynsql.IntColumn("id", "ID").Build()
	minionIDCol := dynsql.IntColumn("minion_id", "终端 ID").Build()
	totalNumCol := dynsql.IntColumn("total_num", "漏洞总数").Build()

	componentTable := dynsql.Builder().
		Filters(idCol, minionIDCol, totalNumCol).
		Build()
	return &sbomREST{
		svc:            svc,
		componentTable: componentTable,
	}
}

type sbomREST struct {
	svc            service.SBOMService
	componentTable dynsql.Table
}

func (rest *sbomREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/sbom/components").GET(rest.PageComponent)
}

func (rest *sbomREST) PageComponent(c *ship.Context) error {
	var req param.PageSQL
	if err := c.BindQuery(&req); err != nil {
		return err
	}
	scope, err := rest.componentTable.Inter(req.Input)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}

	page := req.Pager()
	ctx := c.Request().Context()
	count, dats := rest.svc.PageComponent(ctx, page, scope)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}
