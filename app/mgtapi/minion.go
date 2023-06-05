package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/vela-ssoc/vela-manager/bridge/linkhub"
	"github.com/vela-ssoc/vela-manager/errcode"
	"github.com/xgfone/ship/v5"
)

func Minion(hub linkhub.Huber, svc service.MinionService) route.Router {
	idCol := dynsql.IntColumn("minion.id", "ID").Build()
	tagCol := dynsql.StringColumn("minion_tag.tag", "标签").
		Operators([]dynsql.Operator{dynsql.Eq, dynsql.Like, dynsql.In}).
		Build()
	inetCol := dynsql.StringColumn("minion.inet", "终端IP").Build()
	verCol := dynsql.StringColumn("minion.edition", "版本").Build()
	idcCol := dynsql.StringColumn("minion.idc", "机房").Build()
	ibuCol := dynsql.StringColumn("minion.ibu", "部门").Build()
	commentCol := dynsql.StringColumn("minion.`comment`", "描述").Build()
	statusEnums := dynsql.IntEnum().Set(1, "未激活").Set(2, "离线").
		Set(3, "在线").Set(4, "已删除")
	statusCol := dynsql.IntColumn("minion.status", "状态").Enums(statusEnums).Build()
	goosEnums := dynsql.StringEnum().Sames([]string{"linux", "windows", "darwin"})
	goosCol := dynsql.StringColumn("minion.goos", "操作系统").Enums(goosEnums).Build()
	archEnums := dynsql.StringEnum().Sames([]string{"amd64", "386", "arm64", "arm"})
	archCol := dynsql.StringColumn("minion.arch", "系统架构").Enums(archEnums).Build()
	brkCol := dynsql.StringColumn("minion.broker_name", "代理节点").Build()
	dutyCol := dynsql.StringColumn("minion.op_duty", "运维负责人").Build()
	catCol := dynsql.TimeColumn("minion.created_at", "创建时间").Build()
	upCol := dynsql.TimeColumn("minion.uptime", "上线时间").Build()

	table := dynsql.Builder().
		Filters(tagCol, inetCol, goosCol, archCol, statusCol, verCol, idcCol, ibuCol, commentCol,
			brkCol, dutyCol, catCol, upCol, idCol).
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

func (rest *minionREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/minion/cond").Data(route.Ignore()).GET(rest.Cond)
	bearer.Route("/minions").Data(route.Ignore()).GET(rest.Page)
	bearer.Route("/minion").
		Data(route.Ignore()).GET(rest.Detail).
		Data(route.Named("新增 agent 节点")).POST(rest.Create).
		Data(route.Named("逻辑删除 agent 节点")).DELETE(rest.Delete)
	bearer.Route("/minion/drop").Data(route.Named("物理删除 agent 节点")).DELETE(rest.Drop)
	bearer.Route("/minion/activate").Data(route.Named("激活 agent 节点")).PATCH(rest.Activate)
	bearer.Route("/sheet/minion").Data(route.Ignore()).GET(rest.CSV)
}

func (rest *minionREST) Cond(c *ship.Context) error {
	res := rest.table.Schema()
	return c.JSON(http.StatusOK, res)
}

func (rest *minionREST) Page(c *ship.Context) error {
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

func (rest *minionREST) Detail(c *ship.Context) error {
	var req param.IntID
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	res, err := rest.svc.Detail(ctx, req.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (rest *minionREST) Drop(c *ship.Context) error {
	var req param.IntID
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	err := rest.svc.Drop(ctx, req.ID)

	return err
}

func (rest *minionREST) Create(c *ship.Context) error {
	var req param.MinionCreate
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	err := rest.svc.Create(ctx, &req)

	return err
}

func (rest *minionREST) Delete(c *ship.Context) error {
	var req dynsql.Input
	if err := c.Bind(&req); err != nil {
		return err
	}
	if len(req.Filters) == 0 {
		return errcode.ErrDeleteFailed
	}

	scope, err := rest.table.Inter(req)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}

	ctx := c.Request().Context()

	return rest.svc.Delete(ctx, scope)
}

func (rest *minionREST) Activate(c *ship.Context) error {
	var req param.OptionalIDs
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Activate(ctx, req.ID)
}

func (rest *minionREST) CSV(c *ship.Context) error {
	ctx := c.Request().Context()
	stm := rest.svc.CSV(ctx)

	c.SetRespHeader(ship.HeaderContentDisposition, stm.Disposition())

	return c.Stream(http.StatusOK, stm.MIME(), stm)
}
