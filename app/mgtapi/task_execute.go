package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-common-mb/param/request"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/vela-ssoc/vela-manager/param/mrequest"
	"github.com/xgfone/ship/v5"
)

func NewTaskExecute(svc *service.TaskExecute) *TaskExecute {
	return &TaskExecute{
		svc: svc,
	}
}

type TaskExecute struct {
	svc *service.TaskExecute
}

func (tex *TaskExecute) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/task-executes").Data(route.Ignore()).GET(tex.page)
	bearer.Route("/task-execute").
		Data(route.Ignore()).DELETE(tex.remove).
		Data(route.Ignore()).GET(tex.details)
}

func (tex *TaskExecute) page(c *ship.Context) error {
	req := new(mrequest.TaskExecutePage)
	if err := c.BindQuery(req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	res, err := tex.svc.Page(ctx, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (tex *TaskExecute) remove(c *ship.Context) error {
	req := new(request.Int64ID)
	if err := c.BindQuery(req); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return tex.svc.Remove(ctx, req.ID)
}

func (tex *TaskExecute) details(c *ship.Context) error {
	req := new(request.Int64ID)
	if err := c.BindQuery(req); err != nil {
		return err
	}
	ctx := c.Request().Context()

	ret, err := tex.svc.Details(ctx, req.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}
