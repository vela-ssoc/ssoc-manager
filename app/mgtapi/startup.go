package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-common-mb/param/request"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func NewStartup(svc *service.Startup) *Startup {
	return &Startup{
		svc: svc,
	}
}

type Startup struct {
	svc *service.Startup
}

func (stp *Startup) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/startup").
		Data(route.Ignore()).GET(stp.get).
		Data(route.Named("修改 startup 配置")).POST(stp.update)

	bearer.Route("/startup/fallback").
		Data(route.Ignore()).GET(stp.fallback).
		Data(route.Named("修改 startup 默认配置")).POST(stp.updateFallback)
}

func (stp *Startup) get(c *ship.Context) error {
	req := new(request.Int64ID)
	if err := c.BindQuery(req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	ret, _ := stp.svc.Get(ctx, req.ID)

	return c.JSON(http.StatusOK, ret)
}

func (stp *Startup) update(c *ship.Context) error {
	req := new(param.StartupUpdate)
	if err := c.Bind(req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	err := stp.svc.Update(ctx, req)

	return err
}

func (stp *Startup) fallback(c *ship.Context) error {
	ctx := c.Request().Context()
	ret, _ := stp.svc.Fallback(ctx)

	return c.JSON(http.StatusOK, ret)
}

func (stp *Startup) updateFallback(c *ship.Context) error {
	req := new(param.StartupFallbackUpdate)
	if err := c.Bind(req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	err := stp.svc.UpdateFallback(ctx, req)

	return err
}
