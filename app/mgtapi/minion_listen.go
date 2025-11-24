package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-common-mb/dynsql"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func NewMinionListen(svc *service.MinionListen) *MinionListen {
	inetCol := dynsql.StringColumn("inet", "终端IP").Build()
	protoCol := dynsql.IntColumn("protocol", "协议").Build()
	lportCol := dynsql.IntColumn("local_port", "本地端口").Build()
	pidCol := dynsql.IntColumn("pid", "PID").Build()
	procCol := dynsql.StringColumn("process", "进程名").Build()
	unameCol := dynsql.StringColumn("username", "用户名").Build()
	familyCol := dynsql.IntColumn("family", "family").Build()
	lipCol := dynsql.StringColumn("local_ip", "本地IP").Build()
	pathCol := dynsql.StringColumn("path", "路径").Build()
	midCol := dynsql.StringColumn("minion_id", "节点 ID").Build()

	table := dynsql.Builder().
		Filters(inetCol, protoCol, lportCol, pidCol, procCol, unameCol, familyCol, lipCol, pathCol, midCol).
		Build()

	return &MinionListen{
		svc:   svc,
		table: table,
	}
}

type MinionListen struct {
	svc   *service.MinionListen
	table dynsql.Table
}

func (rest *MinionListen) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/listen/cond").Data(route.Ignore()).GET(rest.Cond)
	bearer.Route("/listens").Data(route.Ignore()).GET(rest.Page)
}

func (rest *MinionListen) Cond(c *ship.Context) error {
	res := rest.table.Schema()
	return c.JSON(http.StatusOK, res)
}

func (rest *MinionListen) Page(c *ship.Context) error {
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
