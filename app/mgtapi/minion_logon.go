package mgtapi

import (
	"net/http"
	"strconv"

	"github.com/vela-ssoc/ssoc-common-mb/dynsql"
	"github.com/vela-ssoc/ssoc-common-mb/param/request"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func NewMinionLogon(svc service.MinionLogonService) *MinionLogon {
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

	return &MinionLogon{
		svc:   svc,
		table: table,
	}
}

type MinionLogon struct {
	svc   service.MinionLogonService
	table dynsql.Table
}

func (rest *MinionLogon) Route(anon, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/logon/cond").Data(route.Ignore()).GET(rest.Cond)
	bearer.Route("/logons").Data(route.Ignore()).GET(rest.Page)
	bearer.Route("/logon/attack").Data(route.Ignore()).POST(rest.Attack)
	anon.Route("/logon/recent").Data(route.Ignore()).GET(rest.Recent)
	bearer.Route("/logon/history").Data(route.Ignore()).GET(rest.History)
	bearer.Route("/logon/ignore").Data(route.Ignore()).PATCH(rest.Ignore)
	bearer.Route("/logon/alert").Data(route.Ignore()).PATCH(rest.Alert)
}

func (rest *MinionLogon) Cond(c *ship.Context) error {
	res := rest.table.Schema()
	return c.JSON(http.StatusOK, res)
}

func (rest *MinionLogon) Page(c *ship.Context) error {
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

func (rest *MinionLogon) Attack(c *ship.Context) error {
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

func (rest *MinionLogon) Recent(c *ship.Context) error {
	day := c.Query("day")
	days, _ := strconv.Atoi(day)
	if days > 30 || days < 1 { // 最多支持30天内查询，参数错误或超过有效范围默认为7天
		days = 7
	}

	ctx := c.Request().Context()
	dats := rest.svc.Recent(ctx, days)

	return c.JSON(http.StatusOK, dats)
}

func (rest *MinionLogon) History(c *ship.Context) error {
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

func (rest *MinionLogon) Alert(*ship.Context) error {
	return nil
}

func (rest *MinionLogon) Ignore(c *ship.Context) error {
	var req request.Int64ID
	if err := c.Bind(&req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	return rest.svc.Ignore(ctx, req.ID)
}
