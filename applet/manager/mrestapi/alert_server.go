package mrestapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-manager/applet/manager/mservice"
	"github.com/xgfone/ship/v5"
)

func NewAlertServer(svc *mservice.AlertServer) *AlertServer {
	return &AlertServer{
		svc: svc,
	}
}

type AlertServer struct {
	svc *mservice.AlertServer
}

func (alt *AlertServer) Route(r *ship.RouteGroupBuilder) error {
	r.Route("/alert-server").
		GET(alt.find).
		POST(alt.upsert)
	return nil
}

func (alt *AlertServer) find(c *ship.Context) error {
	ctx := c.Request().Context()
	dat := alt.svc.Find(ctx)

	return c.JSON(http.StatusOK, dat)
}

func (alt *AlertServer) upsert(c *ship.Context) error {
	return nil
}
