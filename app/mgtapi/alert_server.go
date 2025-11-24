package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/vela-ssoc/ssoc-manager/param/mrequest"
	"github.com/xgfone/ship/v5"
)

func NewAlertServer(svc *service.AlertServer) *AlertServer {
	return &AlertServer{
		svc: svc,
	}
}

type AlertServer struct {
	svc *service.AlertServer
}

func (rest *AlertServer) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/alert-server").
		Data(route.Ignore()).
		GET(rest.find).
		POST(rest.upsert).
		DELETE(rest.delete)
}

func (rest *AlertServer) upsert(c *ship.Context) error {
	req := new(mrequest.AlertServerUpsert)
	if err := c.Bind(req); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return rest.svc.Upsert(ctx, req)
}

func (rest *AlertServer) delete(c *ship.Context) error {
	ctx := c.Request().Context()
	return rest.svc.Delete(ctx)
}

func (rest *AlertServer) find(c *ship.Context) error {
	ctx := c.Request().Context()
	data, err := rest.svc.Find(ctx)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, data)
}
