package restapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-manager/applet/manager/service"
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
