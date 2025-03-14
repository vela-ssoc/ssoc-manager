package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-common-mb/param/request"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/vela-ssoc/ssoc-manager/param/mrequest"
	"github.com/xgfone/ship/v5"
)

func Email(svc service.EmailService) route.Router {
	return &emailREST{
		svc: svc,
	}
}

type emailREST struct {
	svc service.EmailService
}

func (rest *emailREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/emails").Data(route.Ignore()).GET(rest.Page)
	bearer.Route("/email").
		Data(route.Named("新增邮箱配置")).POST(rest.Create).
		Data(route.Named("修改邮箱配置")).PUT(rest.Update).
		Data(route.Named("删除邮箱配置")).DELETE(rest.Delete)
}

func (rest *emailREST) Page(c *ship.Context) error {
	var req param.Page
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	page := req.Pager()
	count, dats := rest.svc.Page(ctx, page)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}

func (rest *emailREST) Create(c *ship.Context) error {
	var req mrequest.EmailCreate
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	err := rest.svc.Create(ctx, &req)

	return err
}

func (rest *emailREST) Update(c *ship.Context) error {
	var req mrequest.EmailUpdate
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	err := rest.svc.Update(ctx, &req)

	return err
}

func (rest *emailREST) Delete(c *ship.Context) error {
	var req request.Int64ID
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	err := rest.svc.Delete(ctx, req.ID)

	return err
}
