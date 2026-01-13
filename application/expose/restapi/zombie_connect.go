package restapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/application/expose/request"
	"github.com/vela-ssoc/ssoc-manager/application/expose/service"
	"github.com/xgfone/ship/v5"
)

type ZombieConnect struct {
	svc *service.ZombieConnect
}

func NewZombieConnect(svc *service.ZombieConnect) *ZombieConnect {
	return &ZombieConnect{
		svc: svc,
	}
}

func (zc *ZombieConnect) BindRoute(rgb *ship.RouteGroupBuilder) error {
	rgb.Route("/zombie/connects").GET(zc.page)

	return nil
}

func (zc *ZombieConnect) page(c *ship.Context) error {
	req := new(request.Pages)
	if err := c.BindQuery(req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	res, err := zc.svc.Page(ctx, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}
