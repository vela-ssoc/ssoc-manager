package mgtapi

import (
	"net"
	"net/http"

	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func Deploy(svc service.DeployService) route.Router {
	return &deployREST{
		svc: svc,
	}
}

type deployREST struct {
	svc service.DeployService
}

func (rest *deployREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/deploy/lan").Data(route.Ignore()).GET(rest.LAN)
}

func (rest *deployREST) LAN(c *ship.Context) error {
	res := &param.DeployLAN{Scheme: "http"}
	r := c.Request()
	if r.TLS != nil {
		res.Scheme = "https"
	}

	ctx := r.Context()
	if addr := rest.svc.LAN(ctx); addr != "" {
		res.Addr = addr
		return c.JSON(http.StatusOK, res)
	}

	val := ctx.Value(http.LocalAddrContextKey)
	if ip, ok := val.(net.Addr); ok {
		res.Addr = ip.String()
	} else {
		res.Addr = r.Host
	}

	return c.JSON(http.StatusOK, res)
}
