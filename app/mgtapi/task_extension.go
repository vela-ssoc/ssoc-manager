package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/vela-ssoc/vela-manager/app/session"
	"github.com/vela-ssoc/vela-manager/param/request"
	"github.com/xgfone/ship/v5"
)

func NewTaskExtension(svc *service.TaskExtension) *TaskExtension {
	return &TaskExtension{
		svc: svc,
	}
}

type TaskExtension struct {
	svc *service.TaskExtension
}

func (tim *TaskExtension) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/task-extensions").Data(route.Ignore()).GET(tim.page)
	// bearer.Route("/task-extension/market").Data(route.Ignore()).POST(tim.fromMarket)
	bearer.Route("/task-extension/filter").Data(route.Ignore()).GET(tim.testFilter)
	bearer.Route("/task-extension/code").
		Data(route.Ignore()).POST(tim.createCode).
		Data(route.Ignore()).PATCH(tim.updateCode)
	bearer.Route("/task-extension/publish").
		Data(route.Ignore()).POST(tim.createPublish).
		Data(route.Ignore()).PATCH(tim.updatePublish)
}

func (tim *TaskExtension) page(c *ship.Context) error {
	req := new(param.Page)
	if err := c.BindQuery(req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	pager := req.Pager()
	cnt, rcd := tim.svc.Page(ctx, pager)
	dat := pager.Result(cnt, rcd)

	return c.JSON(http.StatusOK, dat)
}

func (tim *TaskExtension) fromMarket(c *ship.Context) error {
	req := new(param.TaskExtensionFromMarket)
	if err := c.Bind(req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	cu := session.Cast(c.Any)

	return tim.svc.FromMarket(ctx, req, cu)
}

func (tim *TaskExtension) createCode(c *ship.Context) error {
	req := new(param.TaskExtensionCreateCode)
	if err := c.Bind(req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	cu := session.Cast(c.Any)

	ret, err := tim.svc.CreateCode(ctx, req, cu)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}

func (tim *TaskExtension) updateCode(c *ship.Context) error {
	req := new(param.TaskExtensionUpdateCode)
	if err := c.Bind(req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	cu := session.Cast(c.Any)

	ret, err := tim.svc.UpdateCode(ctx, req, cu)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}

func (tim *TaskExtension) createPublish(c *ship.Context) error {
	req := new(param.TaskExtensionCreatePublish)
	if err := c.Bind(req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	cu := session.Cast(c.Any)

	return tim.svc.CreatePublish(ctx, req, cu)
}

func (tim *TaskExtension) updatePublish(c *ship.Context) error {
	req := new(param.TaskExtensionUpdatePublish)
	if err := c.Bind(req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	cu := session.Cast(c.Any)

	return tim.svc.UpdatePublish(ctx, req, cu)
}

func (tim *TaskExtension) testFilter(c *ship.Context) error {
	req := new(request.Int64ID)
	if err := c.BindQuery(req); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return tim.svc.TestF(ctx, req.ID)
}

func (tim *TaskExtension) History(c *ship.Context) error {
	req := new(request.Int64ID)
	if err := c.BindQuery(req); err != nil {
		return err
	}

	return nil
}
