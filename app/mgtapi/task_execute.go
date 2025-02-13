package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/vela-ssoc/vela-manager/param/request"
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
	bearer.Route("/task-execute/items").Data(route.Ignore()).GET(tex.items)
}

func (tex *TaskExecute) page(c *ship.Context) error {
	req := new(request.TaskExecutePage)
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

func (tex *TaskExecute) items(c *ship.Context) error {
	req := new(request.TaskExecuteItems)
	if err := c.BindQuery(req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	res, err := tex.svc.Items(ctx, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}
