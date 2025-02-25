package brestapi

import (
	"github.com/vela-ssoc/vela-manager/applet/brkmux"
	"github.com/vela-ssoc/vela-manager/applet/broker/bservice"
	"github.com/xgfone/ship/v5"
)

func NewPeer(svc *bservice.Peer) *Peer {
	return &Peer{
		svc: svc,
	}
}

type Peer struct {
	svc *bservice.Peer
}

func (pee *Peer) Route(r *ship.RouteGroupBuilder) error {
	r.Route("/peer/heartbeat").GET(pee.heartbeat)
	return nil
}

func (pee *Peer) heartbeat(c *ship.Context) error {
	ctx := c.Request().Context()
	brk := brkmux.FromContext(ctx)
	if brk == nil {
		return nil
	}
	ident, _ := brk.Info()

	return pee.svc.Heartbeat(ctx, ident)
}
