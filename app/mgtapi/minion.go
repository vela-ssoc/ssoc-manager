package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/vela-ssoc/vela-manager/bridge/linkhub"
	"github.com/xgfone/ship/v5"
)

func Minion(hub linkhub.Huber, svc service.MinionService) route.Router {
	idCol := dynsql.IntColumn("minion.id", "ID").Build()
	tagCol := dynsql.StringColumn("minion_tag.tag", "标签").
		Operators([]dynsql.Operator{dynsql.Eq, dynsql.Like, dynsql.In}).
		Build()
	table := dynsql.Builder().
		Filters(tagCol, idCol).
		Build()

	return &minionREST{
		hub:   hub,
		svc:   svc,
		table: table,
	}
}

type minionREST struct {
	hub   linkhub.Huber
	svc   service.MinionService
	table dynsql.Table
}

func (mon *minionREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/minion/cond").GET(mon.Cond)
	bearer.Route("/minions").GET(mon.Page)
	bearer.Route("/minion").GET(mon.Detail)
}

func (mon *minionREST) Cond(c *ship.Context) error {
	res := mon.table.Schema()
	return c.JSON(http.StatusOK, res)
}

func (mon *minionREST) Page(c *ship.Context) error {
	var req param.PageSQL
	if err := c.BindQuery(&req); err != nil {
		return err
	}
	scope, err := mon.table.Inter(req.Input)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}

	page := req.Pager()
	ctx := c.Request().Context()
	count, dats := mon.svc.Page(ctx, page, scope)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}

func (mon *minionREST) Detail(c *ship.Context) error {
	var req param.IntID
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	res, err := mon.svc.Detail(ctx, req.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}
