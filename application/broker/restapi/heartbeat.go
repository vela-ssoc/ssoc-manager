package restapi

import (
	"github.com/vela-ssoc/ssoc-common/muxserver"
	"github.com/vela-ssoc/ssoc-manager/application/broker/service"
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

func (hb *Heartbeat) RegisterRoute(rgb *ship.RouteGroupBuilder) error {
	rgb.Route("/heartbeat").GET(hb.ping)
	return nil
}

func (hb *Heartbeat) ping(c *ship.Context) error {
	ctx := c.Request().Context()
	peer := muxserver.FromContext(ctx)
	id := peer.ID()

	return hb.svc.Ping(ctx, id)
}
