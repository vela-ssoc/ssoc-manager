package restapi

import (
	"github.com/vela-ssoc/ssoc-common-mb/pubsub"
	"github.com/vela-ssoc/ssoc-common/eventsource"
	"github.com/vela-ssoc/ssoc-manager/appv3/response"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/xgfone/ship/v5"
)

type SSEEvent struct {
	dsp pubsub.Dispatcher
}

func NewSSEEvent(dsp pubsub.Dispatcher) *SSEEvent {
	return &SSEEvent{dsp: dsp}
}

func (se *SSEEvent) RegisterHTTP(g shipx.RouteGroup3) error {
	g.ExposeV3.Group.Route("/sse/event").GET(se.subscribe) // 列表显示

	return nil
}

func (se *SSEEvent) subscribe(c *ship.Context) error {
	w, r := c.Response(), c.Request()
	sse := eventsource.Accept(w, r)
	if sse == nil {
		return ship.ErrUnsupportedMediaType
	}
	defer sse.Close()

	_ = sse.Ping() // 快速响应客户端，避免客户端一直 pending

	sub := response.NewSSEEvent(sse)
	se.dsp.AddSubscriber(sub)
	defer se.dsp.DelSubscriber(sub)

	<-sse.Done()

	return nil
}
