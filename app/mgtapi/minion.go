package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/vela-ssoc/vela-manager/bridge/linkhub"
	"github.com/vela-ssoc/vela-manager/errcode"
	"github.com/xgfone/ship/v5"
	"gorm.io/gen"
	"gorm.io/gen/field"
)

func Minion(hub linkhub.Huber, svc service.MinionService) route.Router {
	const (
		idKey         = "minion.id"
		tagKey        = "minion_tag.tag"
		inetKey       = "minion.inet"
		editionKey    = "minion.edition"
		idcKey        = "minion.idc"
		ibuKey        = "minion.ibu"
		commentKey    = "minion.`comment`"
		statusKey     = "minion.status"
		unloadKey     = "minion.unload"
		goosKey       = "minion.goos"
		archKey       = "minion.arch"
		brokerNameKey = "minion.broker_name"
		opDutyKey     = "minion.op_duty"
		createdAtKey  = "minion.created_at"
		uptimeKey     = "minion.uptime"
	)

	idCol := dynsql.IntColumn(idKey, "ID").Build()
	tagCol := dynsql.StringColumn(tagKey, "标签").
		Operators([]dynsql.Operator{dynsql.Eq, dynsql.Like, dynsql.In}).
		Build()
	inetCol := dynsql.StringColumn(inetKey, "终端IP").Build()
	verCol := dynsql.StringColumn(editionKey, "版本").Build()
	idcCol := dynsql.StringColumn(idcKey, "机房").Build()
	ibuCol := dynsql.StringColumn(ibuKey, "部门").Build()
	commentCol := dynsql.StringColumn(commentKey, "描述").Build()
	statusEnums := dynsql.IntEnum().Set(1, "未激活").Set(2, "离线").
		Set(3, "在线").Set(4, "已删除")
	statusCol := dynsql.IntColumn(statusKey, "状态").
		Enums(statusEnums).
		Operators([]dynsql.Operator{dynsql.Eq, dynsql.Ne, dynsql.In, dynsql.NotIn}).
		Build()
	unloadCol := dynsql.BoolColumn(unloadKey, "静默模式").
		Enums(dynsql.BoolEnum().True("开").False("关")).
		Build()
	goosEnums := dynsql.StringEnum().Sames([]string{"linux", "windows", "darwin"})
	goosCol := dynsql.StringColumn(goosKey, "操作系统").
		Enums(goosEnums).
		Operators([]dynsql.Operator{dynsql.Eq, dynsql.Ne, dynsql.In, dynsql.NotIn}).
		Build()
	archEnums := dynsql.StringEnum().Sames([]string{"amd64", "386", "arm64", "arm"})
	archCol := dynsql.StringColumn(archKey, "系统架构").
		Enums(archEnums).
		Operators([]dynsql.Operator{dynsql.Eq, dynsql.Ne, dynsql.In, dynsql.NotIn}).
		Build()
	brkCol := dynsql.StringColumn(brokerNameKey, "代理节点").Build()
	dutyCol := dynsql.StringColumn(opDutyKey, "运维负责人").Build()
	catCol := dynsql.TimeColumn(createdAtKey, "创建时间").Build()
	upCol := dynsql.TimeColumn(uptimeKey, "上线时间").Build()

	table := dynsql.Builder().
		Filters(tagCol, inetCol, goosCol, archCol, statusCol, unloadCol, verCol,
			idcCol, ibuCol, commentCol, brkCol, dutyCol, catCol, upCol, idCol).
		Build()

	tbl := query.Minion
	likes := map[string]field.String{
		tagKey:        query.MinionTag.Tag,
		inetKey:       tbl.Inet,
		editionKey:    tbl.Edition,
		idcKey:        tbl.IDC,
		ibuKey:        tbl.IBu,
		commentKey:    tbl.Comment,
		goosKey:       tbl.Goos,
		archKey:       tbl.Arch,
		brokerNameKey: tbl.BrokerName,
		opDutyKey:     tbl.OpDuty,
	}

	return &minionREST{
		hub:   hub,
		svc:   svc,
		table: table,
		likes: likes,
	}
}

type minionREST struct {
	hub   linkhub.Huber
	svc   service.MinionService
	table dynsql.Table
	likes map[string]field.String
}

func (rest *minionREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/minion/cond").Data(route.Ignore()).GET(rest.Cond)
	bearer.Route("/minions").Data(route.Ignore()).GET(rest.Page)
	bearer.Route("/minion").
		Data(route.Ignore()).GET(rest.Detail).
		Data(route.Named("新增 agent 节点")).POST(rest.Create).
		Data(route.Named("逻辑删除 agent 节点")).DELETE(rest.Delete)
	bearer.Route("/minion/drop").Data(route.Named("物理删除 agent 节点")).DELETE(rest.Drop)
	bearer.Route("/sheet/minion").Data(route.Ignore()).GET(rest.CSV)
	bearer.Route("/minion/upgrade").Data(route.Named("节点检查更新")).PATCH(rest.Upgrade)
	bearer.Route("/minion/batch").Data(route.Named("批量操作")).PATCH(rest.Batch)
	bearer.Route("/minion/unload").Data(route.Named("静默模式开关")).PATCH(rest.Unload)
	bearer.Route("/minion/batch/tag").Data(route.Named("批量标签管理")).PATCH(rest.BatchTag)
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
	likes := rest.keywordSQL(req.Input, page.Keyword())
	count, dats := rest.svc.Page(ctx, page, scope, likes)
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

func (rest *minionREST) CSV(c *ship.Context) error {
	ctx := c.Request().Context()
	stm := rest.svc.CSV(ctx)

	c.SetRespHeader(ship.HeaderContentDisposition, stm.Disposition())

	return c.Stream(http.StatusOK, stm.MIME(), stm)
}

func (rest *minionREST) Upgrade(c *ship.Context) error {
	var req param.MinionUpgradeRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Upgrade(ctx, req.ID, req.Semver)
}

func (rest *minionREST) Unload(c *ship.Context) error {
	var req param.MinionUnloadRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Unload(ctx, req.ID, req.Unload)
}

func (rest *minionREST) Batch(c *ship.Context) error {
	var req param.MinionBatchRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	keyword := req.Like()
	ctx := c.Request().Context()

	if req.Cmd != "resync" {
		if len(req.Filters) == 0 && keyword == "" {
			return errcode.ErrRequiredFilter
		}
	}

	scope, err := rest.table.Inter(req.Input)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}

	likes := rest.keywordSQL(req.Input, req.Like())

	return rest.svc.Batch(ctx, scope, likes, req.Cmd)
}

func (rest *minionREST) Delete(c *ship.Context) error {
	var req param.MinionDeleteRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	keyword := req.Like()
	ctx := c.Request().Context()
	if len(req.Filters) == 0 && keyword == "" {
		return errcode.ErrRequiredFilter
	}

	scope, err := rest.table.Inter(req.Input)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}

	likes := rest.keywordSQL(req.Input, req.Like())

	return rest.svc.Delete(ctx, scope, likes)
}

func (rest *minionREST) BatchTag(c *ship.Context) error {
	var req param.MinionTagRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	keyword := req.Like()
	ctx := c.Request().Context()
	if len(req.Filters) == 0 && keyword == "" {
		return errcode.ErrRequiredFilter
	}
	scope, err := rest.table.Inter(req.Input)
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}
	if len(req.Deletes) == 0 && len(req.Creates) == 0 {
		return nil
	}
	likes := rest.keywordSQL(req.Input, req.Like())

	return rest.svc.BatchTag(ctx, scope, likes, req.Creates, req.Deletes)
}

func (rest *minionREST) keywordSQL(input dynsql.Input, keyword string) []gen.Condition {
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
