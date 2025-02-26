package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-common-mb/param/request"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func Notifier(svc service.NotifierService) route.Router {
	return &notifierREST{
		svc: svc,
	}
}

type notifierREST struct {
	svc service.NotifierService
}

func (rest *notifierREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/notifiers").Data(route.Ignore()).GET(rest.Page)
	bearer.Route("/notifier").
		Data(route.Named("添加告警人")).POST(rest.Create).
		Data(route.Named("修改告警人")).PUT(rest.Update).
		Data(route.Named("删除告警人")).DELETE(rest.Delete)
}

func (rest *notifierREST) Page(c *ship.Context) error {
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

func (rest *notifierREST) Create(c *ship.Context) error {
	var req param.NotifierCreate
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := req.Validate(); err != nil {
		return ship.ErrBadRequest.New(err)
	}

	ctx := c.Request().Context()
	err := rest.svc.Create(ctx, &req)

	return err
}

func (rest *notifierREST) Update(c *ship.Context) error {
	var req param.NotifierUpdate
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := req.Validate(); err != nil {
		return ship.ErrBadRequest.New(err)
	}

	ctx := c.Request().Context()
	err := rest.svc.Update(ctx, &req)

	return err
}

func (rest *notifierREST) Delete(c *ship.Context) error {
	var req request.Int64ID
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	err := rest.svc.Delete(ctx, req.ID)

	return err
}
