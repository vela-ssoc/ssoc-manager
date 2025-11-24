package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-common-mb/dynsql"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func NewAccount(svc *service.Account) *Account {
	inetCol := dynsql.StringColumn("inet", "终端IP").Build()
	pidCol := dynsql.IntColumn("pid", "PID").Build()
	nameCol := dynsql.StringColumn("name", "进程名称").Build()
	unameCol := dynsql.StringColumn("username", "用户").Build()
	execCol := dynsql.StringColumn("executable", "路径").Build()
	stateCol := dynsql.StringColumn("state", "状态").Build()
	cmdlineCol := dynsql.StringColumn("cmdline", "命令行").Build()
	minionIDCol := dynsql.IntColumn("minion_id", "节点 ID").Build()
	updateAtCol := dynsql.TimeColumn("updated_at", "更新时间").Build()

	table := dynsql.Builder().
		Filters(inetCol, pidCol, nameCol, unameCol, execCol, stateCol, cmdlineCol, minionIDCol, updateAtCol).
		Build()
	return &Account{
		svc:   svc,
		table: table,
	}
}

type Account struct {
	svc   *service.Account
	table dynsql.Table
}

func (rest *Account) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/account/cond").Data(route.Ignore()).GET(rest.Cond)
	bearer.Route("/accounts").Data(route.Ignore()).GET(rest.Page)
}

func (rest *Account) Cond(c *ship.Context) error {
	res := rest.table.Schema()
	return c.JSON(http.StatusOK, res)
}

func (rest *Account) Page(c *ship.Context) error {
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
