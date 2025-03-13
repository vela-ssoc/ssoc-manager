package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/vela-ssoc/ssoc-manager/param/mrequest"
	"github.com/vela-ssoc/vela-common-mb/param/request"
	"github.com/xgfone/ship/v5"
)

func ThirdCustomized(svc service.ThirdCustomizedService) route.Router {
	return &thirdCustomizedREST{
		svc: svc,
	}
}

type thirdCustomizedREST struct {
	svc service.ThirdCustomizedService
}

func (rest *thirdCustomizedREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/third/customizes").Data(route.Ignore()).GET(rest.List)
	bearer.Route("/third/customized").
		Data(route.Named("创建 3rd 标签")).POST(rest.Create).
		Data(route.Named("修改 3rd 标签")).PATCH(rest.Update).
		Data(route.Named("删除 3rd 标签")).DELETE(rest.Delete)
}

func (rest *thirdCustomizedREST) List(c *ship.Context) error {
	ctx := c.Request().Context()
	res := rest.svc.List(ctx)
	return c.JSON(http.StatusOK, res)
}

func (rest *thirdCustomizedREST) Create(c *ship.Context) error {
	var req mrequest.ThirdCustomizedCreate
	if err := c.Bind(&req); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return rest.svc.Create(ctx, &req)
}

func (rest *thirdCustomizedREST) Update(c *ship.Context) error {
	var req mrequest.ThirdCustomizedUpdate
	if err := c.Bind(&req); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return rest.svc.Update(ctx, &req)
}

func (rest *thirdCustomizedREST) Delete(c *ship.Context) error {
	var req request.Int64ID
	if err := c.BindQuery(&req); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return rest.svc.Delete(ctx, req.ID)
}
