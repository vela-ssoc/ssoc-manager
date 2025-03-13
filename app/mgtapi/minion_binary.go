package mgtapi

import (
	"net/http"
	"sync/atomic"

	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/vela-ssoc/ssoc-manager/errcode"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-common-mb/param/request"
	"github.com/xgfone/ship/v5"
)

func MinionBinary(svc *service.MinionBinary) route.Router {
	table := dynsql.Builder().
		Filters(
			dynsql.IntColumn("id", "ID").Build(),
			dynsql.StringColumn("goos", "操作系统").Build(),
			dynsql.StringColumn("arch", "系统架构").Build(),
			dynsql.StringColumn("name", "文件名").Build(),
			dynsql.StringColumn("semver", "版本号").Build(),
			dynsql.StringColumn("changelog", "更新日志").Build(),
			dynsql.StringColumn("customized", "定制版").Build(),
			dynsql.BoolColumn("deprecated", "是否过期").Build(),
			dynsql.BoolColumn("unstable", "是否测试版").Build(),
		).
		Build()

	return &minionBinaryREST{
		svc:   svc,
		table: table,
	}
}

type minionBinaryREST struct {
	svc       *service.MinionBinary
	table     dynsql.Table
	uploading atomic.Bool
}

func (rest *minionBinaryREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/monbin/cond").Data(route.Ignore()).GET(rest.Cond)
	bearer.Route("/monbin/classify").Data(route.Ignore()).GET(rest.Classify)
	bearer.Route("/monbins").Data(route.Ignore()).GET(rest.Page)
	bearer.Route("/monbin/supports").Data(route.Ignore()).GET(rest.supports)
	bearer.Route("/monbin/deprecate").
		Data(route.Named("agent 客户端标记为过期")).PATCH(rest.Deprecate)
	bearer.Route("/monbin").
		Data(route.Ignore()).GET(rest.Download).
		Data(route.IgnoreBody("上传 agent 客户端")).POST(rest.Create).
		Data(route.Named("删除 agent 客户端")).DELETE(rest.Delete).
		Data(route.Named("更新 agent 客户端信息")).PATCH(rest.Update)
	bearer.Route("/monbin/release").
		Data(route.Named("推送升级")).PATCH(rest.Release)
}

func (rest *minionBinaryREST) Cond(c *ship.Context) error {
	res := rest.table.Schema()
	return c.JSON(http.StatusOK, res)
}

func (rest *minionBinaryREST) Classify(c *ship.Context) error {
	ctx := c.Request().Context()
	res, err := rest.svc.Classify(ctx)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (rest *minionBinaryREST) Page(c *ship.Context) error {
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

func (rest *minionBinaryREST) Deprecate(c *ship.Context) error {
	var req request.Int64ID
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Deprecate(ctx, req.ID)
}

func (rest *minionBinaryREST) Delete(c *ship.Context) error {
	var req request.Int64ID
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Delete(ctx, req.ID)
}

func (rest *minionBinaryREST) Create(c *ship.Context) error {
	var req param.NodeBinaryCreate
	if err := c.Bind(&req); err != nil {
		return err
	}
	// 限制只有一个用户操作
	if !rest.uploading.CompareAndSwap(false, true) {
		return errcode.ErrTaskBusy
	}
	defer rest.uploading.Store(false)

	ctx := c.Request().Context()

	return rest.svc.Create(ctx, &req)
}

func (rest *minionBinaryREST) Release(c *ship.Context) error {
	var req request.Int64ID
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Release(ctx, req.ID)
}

// Update 更新发行版信息
func (rest *minionBinaryREST) Update(c *ship.Context) error {
	var req param.MinionBinaryUpdate
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Update(ctx, &req)
}

// Download 更新发行版信息
func (rest *minionBinaryREST) Download(c *ship.Context) error {
	var req request.Int64ID
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	file, err := rest.svc.Download(ctx, req.ID)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer file.Close()

	// 此时的 Content-Length = 原始文件 + 隐藏文件
	c.Header().Set(ship.HeaderContentLength, file.ContentLength())
	c.Header().Set(ship.HeaderContentDisposition, file.Disposition())

	return c.Stream(http.StatusOK, file.ContentType(), file)
}

func (rest *minionBinaryREST) supports(c *ship.Context) error {
	dat := rest.svc.Supports()
	return c.JSON(http.StatusOK, dat)
}
