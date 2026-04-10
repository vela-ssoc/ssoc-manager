package restapi

import (
	"github.com/vela-ssoc/ssoc-manager/application/broker/service"
	"github.com/vela-ssoc/ssoc-manager/bridge/linkhub"
	"github.com/xgfone/ship/v5"
)

type Heartbeat struct {
	svc *service.Heartbeat
}

func NewHeartbeat(svc *service.Heartbeat) *Heartbeat {
	return &Heartbeat{
		svc: svc,
	}
}

func (hb *Heartbeat) Router(rgb *ship.RouteGroupBuilder) {
	rgb.Route("/heartbeat").GET(hb.alive)
}

func (hb *Heartbeat) alive(c *ship.Context) error {
	ctx := c.Request().Context()
	p := linkhub.FromContext(ctx)
	id := p.ID()

	return hb.svc.Alive(ctx, id)
}
