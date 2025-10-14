package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-common-mb/param/request"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func NewNotifier(svc *service.Notifier) *Notifier {
	return &Notifier{
		svc: svc,
	}
}

type Notifier struct {
	svc *service.Notifier
}

func (rest *Notifier) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/notifiers").Data(route.Ignore()).GET(rest.Page)
	bearer.Route("/notifier").
		Data(route.Named("添加告警人")).POST(rest.Create).
		Data(route.Named("修改告警人")).PUT(rest.Update).
		Data(route.Named("删除告警人")).DELETE(rest.Delete)
}

func (rest *Notifier) Page(c *ship.Context) error {
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

func (rest *Notifier) Create(c *ship.Context) error {
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

func (rest *Notifier) Update(c *ship.Context) error {
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

func (rest *Notifier) Delete(c *ship.Context) error {
	var req request.Int64ID
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	err := rest.svc.Delete(ctx, req.ID)

	return err
}
