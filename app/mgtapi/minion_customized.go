package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func MinionCustomized(svc service.MinionCustomizedService) route.Router {
	return &minionCustomizedREST{
		svc: svc,
	}
}

type minionCustomizedREST struct {
	svc service.MinionCustomizedService
}

func (rest *minionCustomizedREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/monbin/customizes").Data(route.Ignore()).GET(rest.List)
	bearer.Route("/monbin/customized").
		Data(route.Named("创建定制版标签")).POST(rest.Create).
		Data(route.Named("删除定制版标签")).DELETE(rest.Delete)
}

func (rest *minionCustomizedREST) List(c *ship.Context) error {
	ctx := c.Request().Context()
	res := rest.svc.List(ctx)
	return c.JSON(http.StatusOK, res)
}

func (rest *minionCustomizedREST) Create(c *ship.Context) error {
	var req param.MinionCustomizedCreate
	if err := c.Bind(&req); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return rest.svc.Create(ctx, &req)
}

func (rest *minionCustomizedREST) Delete(c *ship.Context) error {
	var req param.IntID
	if err := c.Bind(&req); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return rest.svc.Delete(ctx, req.ID)
}
