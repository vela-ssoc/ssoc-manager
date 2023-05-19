package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func MinionTask(svc service.MinionTaskService) route.Router {
	table := dynsql.Builder().Build()
	return &minionTaskREST{
		svc:   svc,
		table: table,
	}
}

type minionTaskREST struct {
	svc   service.MinionTaskService
	table dynsql.Table
}

func (rest *minionTaskREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/tasks").GET(rest.Page)
	bearer.Route("/task").GET(rest.Detail)
	bearer.Route("/task/minion").GET(rest.Minion)
}

func (rest *minionTaskREST) Cond(c *ship.Context) error {
	res := rest.table.Schema()
	return c.JSON(http.StatusOK, res)
}

func (rest *minionTaskREST) Page(c *ship.Context) error {
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

func (rest *minionTaskREST) Detail(c *ship.Context) error {
	var req param.MinionTaskDetailRequest
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	res, err := rest.svc.Detail(ctx, req.ID, req.SubstanceID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (rest *minionTaskREST) Minion(c *ship.Context) error {
	var req param.IntID
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	dats, err := rest.svc.Minion(ctx, req.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dats)
}
