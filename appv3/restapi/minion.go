package restapi

import (
	"errors"
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"github.com/vela-ssoc/ssoc-manager/appv3/service"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/xgfone/ship/v5"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Minion struct {
	svc *service.Minion
}

func NewMinion(svc *service.Minion) *Minion {
	return &Minion{svc: svc}
}

func (mon *Minion) RegisterHTTP(g shipx.RouteGroup) error {
	g.ExposeV3.Group.Route("/minions").GET(mon.page)
	g.ExposeV1.Group.Route("/cmdb").GET(mon.cmdb)

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

func (mon *Minion) cmdb(c *ship.Context) error {
	var param request.Int64ID
	if err := c.BindQuery(&param); err != nil {
		return err
	}

	ctx := c.Request().Context()
	ret, err := mon.svc.CMDB(ctx, param.ID)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}
