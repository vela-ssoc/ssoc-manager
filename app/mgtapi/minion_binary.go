package mgtapi

import (
	"net/http"
	"sync/atomic"

	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/vela-ssoc/vela-manager/errcode"
	"github.com/xgfone/ship/v5"
)

func MinionBinary(svc service.MinionBinaryService) route.Router {
	return &minionBinaryREST{
		svc: svc,
	}
}

type minionBinaryREST struct {
	svc       service.MinionBinaryService
	uploading atomic.Bool
}

func (rest *minionBinaryREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/monbins").Data(route.Ignore()).GET(rest.Page)
	bearer.Route("/monbin/deprecate").
		Data(route.Named("agent 客户端标记为过期")).PATCH(rest.Deprecate)
	bearer.Route("/monbin").
		Data(route.Named("上传 agent 客户端")).POST(rest.Create).
		Data(route.Named("删除 agent 客户端")).DELETE(rest.Delete)
	bearer.Route("/monbin/release").
		Data(route.Named("推送升级")).PATCH(rest.Release)
}

func (rest *minionBinaryREST) Page(c *ship.Context) error {
	var req param.Page
	if err := c.BindQuery(&req); err != nil {
		return err
	}
	page := req.Pager()
	ctx := c.Request().Context()

	count, dats := rest.svc.Page(ctx, page)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}

func (rest *minionBinaryREST) Deprecate(c *ship.Context) error {
	var req param.IntID
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Deprecate(ctx, req.ID)
}

func (rest *minionBinaryREST) Delete(c *ship.Context) error {
	var req param.IntID
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
	var req param.IntID
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Release(ctx, req.ID)
}
