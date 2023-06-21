package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/vela-ssoc/vela-manager/app/session"
	"github.com/xgfone/ship/v5"
)

func Substance(svc service.SubstanceService) route.Router {
	return &substanceREST{
		svc: svc,
	}
}

type substanceREST struct {
	svc service.SubstanceService
}

func (rest *substanceREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/minion/reload").Data(route.Named("重启配置")).PATCH(rest.Reload)
	bearer.Route("/minion/command").Data(route.Named("发送配置指令")).PATCH(rest.Command)
	bearer.Route("/substances").Data(route.Ignore()).GET(rest.Page)
	bearer.Route("/substance/indices").Data(route.Ignore()).GET(rest.Indices)
	bearer.Route("/substance").
		Data(route.Ignore()).GET(rest.Detail).
		Data(route.Named("新增配置")).POST(rest.Create).
		Data(route.Named("修改配置")).PUT(rest.Update).
		Data(route.Named("删除配置")).DELETE(rest.Delete)
}

func (rest *substanceREST) Indices(c *ship.Context) error {
	var req param.Index
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	idx := req.Indexer()
	ctx := c.Request().Context()
	dats := rest.svc.Indices(ctx, idx)

	return c.JSON(http.StatusOK, dats)
}

func (rest *substanceREST) Page(c *ship.Context) error {
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

func (rest *substanceREST) Detail(c *ship.Context) error {
	var req param.IntID
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	id := req.ID
	res, err := rest.svc.Detail(ctx, id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (rest *substanceREST) Create(c *ship.Context) error {
	var req param.SubstanceCreate
	if err := c.Bind(&req); err != nil {
		return err
	}

	cu := session.Cast(c.Any)
	ctx := c.Request().Context()

	return rest.svc.Create(ctx, &req, cu.ID)
}

func (rest *substanceREST) Update(c *ship.Context) error {
	var req param.SubstanceUpdate
	if err := c.Bind(&req); err != nil {
		return err
	}

	cu := session.Cast(c.Any)
	ctx := c.Request().Context()

	return rest.svc.Update(ctx, &req, cu.ID)
}

func (rest *substanceREST) Delete(c *ship.Context) error {
	var req param.IntID
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Delete(ctx, req.ID)
}

// Reload 重新加载指定节点上的指定配置
func (rest *substanceREST) Reload(c *ship.Context) error {
	var req param.SubstanceReload
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Reload(ctx, req.ID, req.SubstanceID)
}

// Command 重新加载指定节点上的指定配置
func (rest *substanceREST) Command(c *ship.Context) error {
	var req param.SubstanceCommand
	if err := c.Bind(&req); err != nil {
		return err
	}

	mid := req.ID
	ctx := c.Request().Context()

	switch req.Cmd {
	case "resync":
		return rest.svc.Resync(ctx, mid)
	default:
		return rest.svc.Command(ctx, mid, req.Cmd)
	}
}
