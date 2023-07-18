package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
	"gorm.io/gen"
	"gorm.io/gen/field"
)

func SubstanceTask(svc service.SubstanceTaskService) route.Router {
	inetKey := "inet"
	bnameKey := "broker_name"
	reasonKey := "reason"

	tbl := dynsql.Builder().
		Filters(
			dynsql.StringColumn(inetKey, "节点IP").Build(),
			dynsql.StringColumn(reasonKey, "错误原因").Build(),
			dynsql.StringColumn(bnameKey, "代理节点名称").Build(),
			dynsql.BoolColumn("executed", "已执行").Enums(dynsql.BoolEnum().True("是").False("否")).Build(),
			dynsql.BoolColumn("failed", "执行失败").Enums(dynsql.BoolEnum().True("是")).Build(),
			// dynsql.IntColumn("task_id", "任务ID").Build(),
			dynsql.IntColumn("minion_id", "节点ID").Build(),
			dynsql.IntColumn("broker_id", "代理节点ID").Build(),
		).
		Build()

	dao := query.SubstanceTask
	likes := map[string]field.String{
		inetKey:   dao.Inet,
		bnameKey:  dao.BrokerName,
		reasonKey: dao.Reason,
	}

	return &substanceTaskREST{
		svc:   svc,
		tbl:   tbl,
		likes: likes,
	}
}

type substanceTaskREST struct {
	svc   service.SubstanceTaskService
	tbl   dynsql.Table
	likes map[string]field.String
}

func (rest *substanceTaskREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/effect/progress/cond").Data(route.Ignore()).GET(rest.Cond)
	bearer.Route("/effect/progresses").Data(route.Ignore()).GET(rest.Page)
	bearer.Route("/effect/progress/histories").Data(route.Ignore()).GET(rest.Histories)
}

func (rest *substanceTaskREST) Cond(c *ship.Context) error {
	res := rest.tbl.Schema()
	return c.JSON(http.StatusOK, res)
}

func (rest *substanceTaskREST) Page(c *ship.Context) error {
	var req param.IDPageSQL
	if err := c.BindQuery(&req); err != nil {
		return err
	}
	scope, err := rest.tbl.Inter(req.Input)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}
	page := req.Pager()
	ctx := c.Request().Context()
	likes := rest.keywordSQL(req.Input, page.Keyword())
	count, dats := rest.svc.Page(ctx, req.ID, page, scope, likes)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}

func (rest *substanceTaskREST) keywordSQL(input dynsql.Input, keyword string) []gen.Condition {
	if keyword == "" {
		return nil
	}

	hm := make(map[string]struct{}, len(input.Filters))
	for _, fl := range input.Filters {
		hm[fl.Col] = struct{}{}
	}

	ret := make([]gen.Condition, 0, len(rest.likes))
	for k, f := range rest.likes {
		if _, ok := hm[k]; ok {
			continue
		}
		ret = append(ret, f.Like(keyword))
	}

	return ret
}

func (rest *substanceTaskREST) Histories(c *ship.Context) error {
	var req param.PageSQL
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	scope, err := rest.tbl.Inter(req.Input)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}
	page := req.Pager()
	ctx := c.Request().Context()
	likes := rest.keywordSQL(req.Input, page.Keyword())
	count, dats := rest.svc.Histories(ctx, page, scope, likes)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}
