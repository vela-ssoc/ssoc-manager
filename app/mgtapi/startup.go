package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-common-mb-itai/dal/model"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func Startup(svc service.StartupService) route.Router {
	return &startupREST{
		svc: svc,
	}
}

type startupREST struct {
	svc service.StartupService
}

func (rest *startupREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/startup").
		Data(route.Ignore()).GET(rest.Detail).
		Data(route.Named("修改 startup 配置")).PUT(rest.Update)
}

func (rest *startupREST) Detail(c *ship.Context) error {
	var req param.IntID
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	dat, err := rest.svc.Detail(ctx, req.ID)
	if err != nil {
		return err
	}
	res := &param.StartupDetail{Param: dat}

	return c.JSON(http.StatusOK, res)
}

func (rest *startupREST) Update(c *ship.Context) error {
	var req model.Startup
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	err := rest.svc.Update(ctx, &req)

	return err
}
