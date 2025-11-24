package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-common-mb/param/request"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/vela-ssoc/ssoc-manager/app/session"
	"github.com/xgfone/ship/v5"
)

func NewSubstance(svc *service.Substance) *Substance {
	return &Substance{
		svc: svc,
	}
}

type Substance struct {
	svc *service.Substance
}

func (sst *Substance) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/minion/reload").Data(route.Named("重启配置")).PATCH(sst.Reload)
	bearer.Route("/minion/command").Data(route.Named("发送配置指令")).PATCH(sst.Command)
	bearer.Route("/substances").Data(route.Ignore()).GET(sst.Page)
	bearer.Route("/substance/indices").Data(route.Ignore()).GET(sst.Indices)
	bearer.Route("/substance").
		Data(route.Ignore()).GET(sst.Detail).
		Data(route.Named("新增配置")).POST(sst.Create).
		Data(route.Named("修改配置")).PUT(sst.Update).
		Data(route.Named("删除配置")).DELETE(sst.Delete)

	bearer.Route("/substance/exclude").
		Data(route.Named("节点排除配置")).POST(sst.exclude)
	bearer.Route("/substance/unexclude").
		Data(route.Named("节点取消排除配置")).POST(sst.unexclude)
}

func (sst *Substance) Indices(c *ship.Context) error {
	var req param.Index
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	idx := req.Indexer()
	ctx := c.Request().Context()
	dats := sst.svc.Indices(ctx, idx)

	return c.JSON(http.StatusOK, dats)
}

func (sst *Substance) Page(c *ship.Context) error {
	var req param.Page
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	page := req.Pager()
	ctx := c.Request().Context()
	count, dats := sst.svc.Page(ctx, page)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}

func (sst *Substance) Detail(c *ship.Context) error {
	var req request.Int64ID
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	id := req.ID
	res, err := sst.svc.Detail(ctx, id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (sst *Substance) Create(c *ship.Context) error {
	var req param.SubstanceCreate
	if err := c.Bind(&req); err != nil {
		return err
	}

	cu := session.Cast(c.Any)
	ctx := c.Request().Context()

	return sst.svc.Create(ctx, &req, cu.ID)
}

func (sst *Substance) Update(c *ship.Context) error {
	var req param.SubstanceUpdate
	if err := c.Bind(&req); err != nil {
		return err
	}

	cu := session.Cast(c.Any)
	ctx := c.Request().Context()

	tid, err := sst.svc.Update(ctx, &req, cu.ID)
	if err != nil {
		return err
	}
	res := &request.Int64ID{ID: tid}

	return c.JSON(http.StatusOK, res)
}

func (sst *Substance) Delete(c *ship.Context) error {
	var req request.Int64ID
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return sst.svc.Delete(ctx, req.ID)
}

// Reload 重新加载指定节点上的指定配置
func (sst *Substance) Reload(c *ship.Context) error {
	var req param.SubstanceReload
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return sst.svc.Reload(ctx, req.ID, req.SubstanceID)
}

// Command 重新加载指定节点上的指定配置
func (sst *Substance) Command(c *ship.Context) error {
	var req param.SubstanceCommand
	if err := c.Bind(&req); err != nil {
		return err
	}

	mid := req.ID
	ctx := c.Request().Context()

	switch req.Cmd {
	case "resync":
		return sst.svc.Resync(ctx, mid)
	default:
		return sst.svc.Command(ctx, mid, req.Cmd)
	}
}

func (sst *Substance) exclude(c *ship.Context) error {
	req := new(request.SubstanceExclude)
	if err := c.Bind(req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	mid, sid := req.MinionID, req.SubstanceID

	return sst.svc.Exclude(ctx, mid, sid)
}

func (sst *Substance) unexclude(c *ship.Context) error {
	req := new(request.SubstanceExclude)
	if err := c.Bind(req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	mid, sid := req.MinionID, req.SubstanceID

	return sst.svc.Unexclude(ctx, mid, sid)
}
