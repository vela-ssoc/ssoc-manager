package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/param/request"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func NewStartup(svc service.StartupService) *Startup {
	return &Startup{
		svc: svc,
	}
}

type Startup struct {
	svc service.StartupService
}

func (rest *Startup) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/startup").
		Data(route.Ignore()).GET(rest.Detail).
		Data(route.Named("修改 startup 配置")).PUT(rest.Update)
}

func (rest *Startup) Detail(c *ship.Context) error {
	var req request.Int64ID
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

func (rest *Startup) Update(c *ship.Context) error {
	var req model.Startup
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	err := rest.svc.Update(ctx, &req)

	return err
}
