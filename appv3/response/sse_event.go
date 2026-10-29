package response

import (
	"github.com/vela-ssoc/ssoc-common-mb/eventsource"
	"github.com/vela-ssoc/ssoc-common-mb/pubsub"
)

type sseEvent struct {
	sse eventsource.EventSource
}

func NewSSEEvent(sse eventsource.EventSource) pubsub.Subscriber {
	return &sseEvent{sse: sse}
}

func (s *sseEvent) OnEvent(event string, data any) error {
	return s.sse.JSON(event, data)
}
