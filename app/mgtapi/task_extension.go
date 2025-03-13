package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/vela-ssoc/ssoc-manager/app/session"
	"github.com/vela-ssoc/ssoc-manager/param/mrequest"
	"github.com/vela-ssoc/vela-common-mb/param/request"
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
	bearer.Route("/task-extension").Data(route.Ignore()).DELETE(tim.remove)
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

func (tim *TaskExtension) createCode(c *ship.Context) error {
	req := new(mrequest.TaskExtensionCreateCode)
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
	req := new(mrequest.TaskExtensionUpdateCode)
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
	req := new(mrequest.TaskExtensionCreatePublish)
	if err := c.Bind(req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	cu := session.Cast(c.Any)

	return tim.svc.CreatePublish(ctx, req, cu)
}

func (tim *TaskExtension) updatePublish(c *ship.Context) error {
	req := new(mrequest.TaskExtensionUpdatePublish)
	if err := c.Bind(req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	cu := session.Cast(c.Any)

	return tim.svc.UpdatePublish(ctx, req, cu)
}

func (tim *TaskExtension) remove(c *ship.Context) error {
	req := new(request.Int64ID)
	if err := c.BindQuery(req); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return tim.svc.Delete(ctx, req.ID)
}
