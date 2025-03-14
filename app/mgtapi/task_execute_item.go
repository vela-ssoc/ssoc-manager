package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-common-mb/param/request"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/vela-ssoc/ssoc-manager/param/mrequest"
	"github.com/xgfone/ship/v5"
)

func NewTaskExecuteItem(svc *service.TaskExecuteItem) *TaskExecuteItem {
	return &TaskExecuteItem{svc: svc}
}

type TaskExecuteItem struct {
	svc *service.TaskExecuteItem
}

func (tex *TaskExecuteItem) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/task-execute/items").Data(route.Ignore()).GET(tex.page)
	bearer.Route("/task-execute/item/cond").Data(route.Ignore()).GET(tex.cond)
	bearer.Route("/task-execute/item/codes").Data(route.Ignore()).GET(tex.codeCount)
}

func (tex *TaskExecuteItem) cond(c *ship.Context) error {
	ret := tex.svc.Cond()
	return c.JSON(http.StatusOK, ret)
}

func (tex *TaskExecuteItem) page(c *ship.Context) error {
	req := new(request.PageConditions)
	if err := c.BindQuery(req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	ret, err := tex.svc.Page(ctx, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}

func (tex *TaskExecuteItem) codeCount(c *ship.Context) error {
	req := new(mrequest.ExecID)
	if err := c.BindQuery(req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	ret, err := tex.svc.CodeCounts(ctx, req.ExecID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}
