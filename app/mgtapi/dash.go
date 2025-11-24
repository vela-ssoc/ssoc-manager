package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func NewDash(svc *service.Dash) *Dash {
	return &Dash{
		svc: svc,
	}
}

type Dash struct {
	svc *service.Dash
}

func (rest *Dash) Route(anon, bearer, _ *ship.RouteGroupBuilder) {
	anon.Route("/dash/status").Data(route.Ignore()).GET(rest.Status)
	anon.Route("/dash/goos").Data(route.Ignore()).GET(rest.Goos)
	bearer.Route("/dash/edition").Data(route.Ignore()).GET(rest.Edition)
	bearer.Route("/dash/evtlvl").Data(route.Ignore()).GET(rest.Evtlvl)
	anon.Route("/dash/risklvl").Data(route.Ignore()).GET(rest.Risklvl)
	bearer.Route("/dash/risksts").Data(route.Ignore()).GET(rest.Risksts)
	bearer.Route("/dash/broker/goos").Data(route.Ignore()).GET(rest.BGoos)
}

func (rest *Dash) Status(c *ship.Context) error {
	ctx := c.Request().Context()
	res := rest.svc.Status(ctx)
	return c.JSON(http.StatusOK, res)
}

func (rest *Dash) Goos(c *ship.Context) error {
	ctx := c.Request().Context()
	res := rest.svc.Goos(ctx)
	return c.JSON(http.StatusOK, res)
}

func (rest *Dash) Edition(c *ship.Context) error {
	ctx := c.Request().Context()
	dats := rest.svc.Edition(ctx)
	res := param.WarpData(dats)

	return c.JSON(http.StatusOK, res)
}

func (rest *Dash) Evtlvl(c *ship.Context) error {
	ctx := c.Request().Context()
	res := rest.svc.Evtlvl(ctx)
	return c.JSON(http.StatusOK, res)
}

// Risklvl 通过 level 维度统计 risk
func (rest *Dash) Risklvl(c *ship.Context) error {
	ctx := c.Request().Context()
	res := rest.svc.Risklvl(ctx)
	return c.JSON(http.StatusOK, res)
}

// Risksts 通过 status 维度统计 risk
func (rest *Dash) Risksts(c *ship.Context) error {
	ctx := c.Request().Context()
	res := rest.svc.Risksts(ctx)
	return c.JSON(http.StatusOK, res)
}

// BGoos 通过 status 维度统计 risk
func (rest *Dash) BGoos(c *ship.Context) error {
	var req param.Page
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, nil)
}
