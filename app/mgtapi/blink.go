package mgtapi

import (
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/bridge/blink"
	"github.com/vela-ssoc/vela-manager/param/mrequest"
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
	anon.Route("/broker/alert").Data(route.IgnoreBody("发送告警")).POST(lnk.Alert)
}

func (lnk *blinkREST) Join(c *ship.Context) error {
	w, r := c.Response(), c.Request()
	lnk.brk.ServeHTTP(w, r)
	return nil
}

func (lnk *blinkREST) Alert(c *ship.Context) error {
	req := new(mrequest.BlinkAlert)
	if err := c.Bind(req); err != nil {
		return err
	}

	// 判断是否是卡片消息

	return nil
}
