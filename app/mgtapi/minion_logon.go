package mgtapi

import (
	"net/http"
	"strconv"

	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func MinionLogon(svc service.MinionLogonService) route.Router {
	msgEnums := []string{"登录成功", "登录失败", "用户注销"}
	inetCol := dynsql.StringColumn("inet", "终端IP").Build()
	userCol := dynsql.StringColumn("user", "账户").Build()
	msgCol := dynsql.StringColumn("msg", "描述").Enums(dynsql.StringEnum().Sames(msgEnums)).Build()
	addrCol := dynsql.StringColumn("addr", "登录地址").Build()
	logonAtCol := dynsql.TimeColumn("logon_at", "登录时间").Build()
	typeCol := dynsql.StringColumn("type", "登录类型").Build()
	midCol := dynsql.StringColumn("minion_id", "节点ID").Build()
	devCol := dynsql.StringColumn("device", "登录设备").Build()
	table := dynsql.Builder().
		Filters(inetCol, userCol, msgCol, addrCol, logonAtCol, typeCol, midCol, devCol).
		Build()

	return &minionLogonREST{
		svc:   svc,
		table: table,
	}
}

type minionLogonREST struct {
	svc   service.MinionLogonService
	table dynsql.Table
}

func (rest *minionLogonREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/logon/cond").GET(rest.Cond)
	bearer.Route("/logons").GET(rest.Page)
	bearer.Route("/logon/attack").POST(rest.Attack)
	bearer.Route("/logon/recent").GET(rest.Recent)
	bearer.Route("/logon/history").GET(rest.History)
}

func (rest *minionLogonREST) Cond(c *ship.Context) error {
	res := rest.table.Schema()
	return c.JSON(http.StatusOK, res)
}

func (rest *minionLogonREST) Page(c *ship.Context) error {
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

func (rest *minionLogonREST) Attack(c *ship.Context) error {
	var req param.PageSQL
	if err := c.Bind(&req); err != nil {
		return err
	}
	scope, err := rest.table.Inter(req.Input)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}
	page := req.Pager()
	ctx := c.Request().Context()

	count, dats := rest.svc.Attack(ctx, page, scope)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}

func (rest *minionLogonREST) Recent(c *ship.Context) error {
	day := c.Query("day")
	days, _ := strconv.Atoi(day)
	if days > 30 || days < 1 { // 最多支持30天内查询，参数错误或超过有效范围默认为7天
		days = 7
	}

	ctx := c.Request().Context()
	dats := rest.svc.Recent(ctx, days)

	return c.JSON(http.StatusOK, dats)
}

func (rest *minionLogonREST) History(c *ship.Context) error {
	var req param.MinionLogonHistory
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	page := req.Pager()
	ctx := c.Request().Context()
	count, dats := rest.svc.History(ctx, page, req.MinionID, req.Name)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}
