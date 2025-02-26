package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/vela-ssoc/vela-manager/param/mrequest"
	"github.com/xgfone/ship/v5"
)

func NewAlertServer(svc *service.AlertServer) route.Router {
	return &alertServerREST{
		svc: svc,
	}
}

type alertServerREST struct {
	svc *service.AlertServer
}

func (rest *alertServerREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/alert-server").
		Data(route.Ignore()).
		GET(rest.find).
		POST(rest.upsert).
		DELETE(rest.delete)
}

func (rest *alertServerREST) upsert(c *ship.Context) error {
	req := new(mrequest.AlertServerUpsert)
	if err := c.Bind(req); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return rest.svc.Upsert(ctx, req)
}

func (rest *alertServerREST) delete(c *ship.Context) error {
	ctx := c.Request().Context()
	return rest.svc.Delete(ctx)
}

func (rest *alertServerREST) find(c *ship.Context) error {
	ctx := c.Request().Context()
	data, err := rest.svc.Find(ctx)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, data)
}
