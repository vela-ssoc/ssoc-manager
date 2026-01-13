package restapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/application/expose/request"
	"github.com/vela-ssoc/ssoc-manager/application/expose/service"
	"github.com/xgfone/ship/v5"
)

type Occupy struct {
	svc *service.Occupy
}

func NewOccupy(svc *service.Occupy) *Occupy {
	return &Occupy{
		svc: svc,
	}
}

func (occ *Occupy) BindRoute(rgb *ship.RouteGroupBuilder) error {
	rgb.Route("/occupy/events").GET(occ.events)
	rgb.Route("/occupy/risks").GET(occ.risks)

	return nil
}

func (occ *Occupy) events(c *ship.Context) error {
	req := new(request.OccupyPages)
	if err := c.BindQuery(req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	res, err := occ.svc.Events(ctx, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (occ *Occupy) risks(c *ship.Context) error {
	req := new(request.OccupyPages)
	if err := c.BindQuery(req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	res, err := occ.svc.Risks(ctx, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}
