package mgtapi

import (
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/bridge/blink"
	"github.com/xgfone/ship/v5"
)

func Blink(brk blink.Handler) route.Router {
	return &blinkREST{brk: brk}
}

type blinkREST struct {
	brk blink.Handler
}

func (lnk *blinkREST) Route(anon, _, _ *ship.RouteGroupBuilder) {
	anon.Route("/broker").Data(route.IgnoreBody("broker 节点上下线")).CONNECT(lnk.Join)
}

func (lnk *blinkREST) Join(c *ship.Context) error {
	w, r := c.Response(), c.Request()
	lnk.brk.ServeHTTP(w, r)
	return nil
}
