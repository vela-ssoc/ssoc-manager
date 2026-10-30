package restapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"github.com/vela-ssoc/ssoc-manager/appv3/service"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/xgfone/ship/v5"
)

type Minion struct {
	svc *service.Minion
}

func NewMinion(svc *service.Minion) *Minion {
	return &Minion{svc: svc}
}

func (mon *Minion) RegisterHTTP(g shipx.RouteGroup3) error {
	g.ExposeV3.Group.Route("/minions").GET(mon.page)

	return nil
}

func (mon *Minion) page(c *ship.Context) error {
	var param request.MinionPage
	if err := c.BindQuery(&param); err != nil {
		return err
	}

	ctx := c.Request().Context()
	ret, err := mon.svc.Page(ctx, param)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}
