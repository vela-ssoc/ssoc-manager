package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-common-mb/param/request"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func NewMinionCustomized(svc *service.MinionCustomized) *MinionCustomized {
	return &MinionCustomized{
		svc: svc,
	}
}

type MinionCustomized struct {
	svc *service.MinionCustomized
}

func (rest *MinionCustomized) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/monbin/customizes").Data(route.Ignore()).GET(rest.List)
	bearer.Route("/monbin/customized").
		Data(route.Named("创建定制版标签")).POST(rest.Create).
		Data(route.Named("删除定制版标签")).DELETE(rest.Delete)
}

func (rest *MinionCustomized) List(c *ship.Context) error {
	ctx := c.Request().Context()
	res := rest.svc.List(ctx)
	return c.JSON(http.StatusOK, res)
}

func (rest *MinionCustomized) Create(c *ship.Context) error {
	var req param.MinionCustomizedCreate
	if err := c.Bind(&req); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return rest.svc.Create(ctx, &req)
}

func (rest *MinionCustomized) Delete(c *ship.Context) error {
	var req request.Int64ID
	if err := c.Bind(&req); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return rest.svc.Delete(ctx, req.ID)
}
