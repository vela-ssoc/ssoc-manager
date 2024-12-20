package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func NewSIEMServer(svc *service.SIEMServer) route.Router {
	return &siemServerREST{
		svc: svc,
	}
}

type siemServerREST struct {
	svc *service.SIEMServer
}

func (rest *siemServerREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/siem-server").
		Data(route.Ignore()).
		GET(rest.find).
		POST(rest.upsert).
		DELETE(rest.delete)
}

func (rest *siemServerREST) upsert(c *ship.Context) error {
	req := new(param.SIEMServerUpsert)
	if err := c.Bind(req); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return rest.svc.Upsert(ctx, req)
}

func (rest *siemServerREST) delete(c *ship.Context) error {
	ctx := c.Request().Context()
	return rest.svc.Delete(ctx)
}

func (rest *siemServerREST) find(c *ship.Context) error {
	ctx := c.Request().Context()
	data, err := rest.svc.Find(ctx)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, data)
}
