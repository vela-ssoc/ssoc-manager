package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func Dash(svc service.DashService) route.Router {
	return &dashREST{
		svc: svc,
	}
}

type dashREST struct {
	svc service.DashService
}

func (rest *dashREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/dash/status").GET(rest.Status)
	bearer.Route("/dash/goos").GET(rest.Goos)
	bearer.Route("/dash/edition").GET(rest.Edition)
	bearer.Route("/dash/evtlvl").GET(rest.Evtlvl)
	bearer.Route("/dash/risklvl").GET(rest.Risklvl)
	bearer.Route("/dash/risksts").GET(rest.Risksts)
}

func (rest *dashREST) Status(c *ship.Context) error {
	ctx := c.Request().Context()
	res := rest.svc.Status(ctx)
	return c.JSON(http.StatusOK, res)
}

func (rest *dashREST) Goos(c *ship.Context) error {
	ctx := c.Request().Context()
	res := rest.svc.Goos(ctx)
	return c.JSON(http.StatusOK, res)
}

func (rest *dashREST) Edition(c *ship.Context) error {
	ctx := c.Request().Context()
	dats := rest.svc.Edition(ctx)
	res := param.WarpData(dats)

	return c.JSON(http.StatusOK, res)
}

func (rest *dashREST) Evtlvl(c *ship.Context) error {
	ctx := c.Request().Context()
	res := rest.svc.Evtlvl(ctx)
	return c.JSON(http.StatusOK, res)
}

// Risklvl 通过 level 维度统计 risk
func (rest *dashREST) Risklvl(c *ship.Context) error {
	ctx := c.Request().Context()
	res := rest.svc.Risklvl(ctx)
	return c.JSON(http.StatusOK, res)
}

// Risksts 通过 status 维度统计 risk
func (rest *dashREST) Risksts(c *ship.Context) error {
	ctx := c.Request().Context()
	res := rest.svc.Risksts(ctx)
	return c.JSON(http.StatusOK, res)
}
