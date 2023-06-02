package mgtapi

import (
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func MinionListen(svc service.MinionListenService) route.Router {
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

	return &minionListenREST{
		svc:   svc,
		table: table,
	}
}

type minionListenREST struct {
	svc   service.MinionListenService
	table dynsql.Table
}

func (rest *minionListenREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/listen/cond").GET(rest.Cond)
	bearer.Route("/listens").GET(rest.Page)
}

func (rest *minionListenREST) Cond(c *ship.Context) error {
	return nil
}

func (rest *minionListenREST) Page(c *ship.Context) error {
	return nil
}
